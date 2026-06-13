import json
import time
import redis
import threading
from datetime import datetime
from collections import defaultdict

r = redis.Redis(host="127.0.0.1", port=6379, decode_responses=True)

user_profiles = defaultdict(lambda: {
    "total_sent": 0.0,
    "transfer_count": 0,
    "avg_amount": 0.0,
    "unique_recipients": set(),
    "first_seen": None,
    "last_transfer": None,
    "hourly_distribution": [0] * 24,
})

ACCOUNT_RISK_CACHE = {}
CACHE_TTL = 300


def analyze_transfer(req: dict) -> dict:
    user_id = req.get("user_id", "")
    amount = float(req.get("amount", 0))
    from_acc = req.get("from_account", "")
    to_acc = req.get("to_account", "")
    req_id = req.get("id", "")

    profile = user_profiles[user_id]
    now = datetime.now()

    if profile["first_seen"] is None:
        profile["first_seen"] = now

    risk_score = 0
    reasons = []

    # 1. New account risk: account created < 24h ago
    cache_key = f"account:created:{from_acc}"
    account_age_hours = None
    created_ts = r.get(cache_key)
    if created_ts:
        age_secs = time.time() - float(created_ts)
        account_age_hours = age_secs / 3600
        if account_age_hours < 1:
            risk_score += 40
            reasons.append(f"Very new account ({account_age_hours*60:.0f}min old)")
        elif account_age_hours < 24:
            risk_score += 15
            reasons.append(f"New account ({account_age_hours:.1f}h old)")

    # 2. Unusual hour (2am-5am)
    hour = now.hour
    profile["hourly_distribution"][hour] += 1
    if 2 <= hour <= 5:
        risk_score += 20
        reasons.append(f"Unusual hour ({hour}:00)")

    # 3. Amount vs average
    if profile["transfer_count"] > 5:
        deviation = abs(amount - profile["avg_amount"]) / max(profile["avg_amount"], 1)
        if deviation > 5:
            risk_score += 25
            reasons.append(f"Amount {deviation:.0f}x different from average")
        elif deviation > 3:
            risk_score += 10
            reasons.append(f"Amount {deviation:.0f}x different from average")

    # 4. Many unique recipients (structuring / layering)
    profile["unique_recipients"].add(to_acc)
    if len(profile["unique_recipients"]) > 15:
        risk_score += 30
        reasons.append(f"Sent to {len(profile['unique_recipients'])} unique accounts")
    elif len(profile["unique_recipients"]) > 8:
        risk_score += 10
        reasons.append(f"Sent to {len(profile['unique_recipients'])} unique accounts")

    # 5. Rapid succession (multiple transfers within 60s)
    last = profile.get("last_transfer")
    if last:
        delta = (now - last).total_seconds()
        if delta < 10:
            risk_score += 35
            reasons.append(f"Transfer within {delta:.0f}s of previous")
        elif delta < 60:
            risk_score += 15
            reasons.append(f"Transfer within {delta:.0f}s of previous")

    # 6. Large round amounts (structuring)
    if amount >= 10000 and amount == int(amount):
        risk_score += 20
        reasons.append("Large round amount (possible structuring)")

    # 7. Accumulating pattern: multiple transfers to same recipient in short time
    recipient_key = f"antifraud:recipient_hits:{to_acc}"
    r.incr(recipient_key)
    r.expire(recipient_key, 3600)
    hits = int(r.get(recipient_key) or 0)
    if hits > 20:
        risk_score += 25
        reasons.append(f"Recipient received {hits} transfers in 1h")

    # Update profile
    profile["transfer_count"] += 1
    profile["total_sent"] += amount
    profile["avg_amount"] = profile["total_sent"] / profile["transfer_count"]
    profile["last_transfer"] = now

    # Verdict
    approved = risk_score < 60
    verdict = "Approved"
    if risk_score >= 80:
        verdict = "Blocked — high risk"
    elif risk_score >= 60:
        verdict = "Blocked — suspicious pattern"

    result = {
        "id": req_id,
        "approved": approved,
        "reason": "; ".join(reasons) if reasons else "No flags",
        "risk_score": min(risk_score, 100),
        "verdict": verdict,
        "engine": "python",
        "user_stats": {
            "transfer_count": profile["transfer_count"],
            "avg_amount": round(profile["avg_amount"], 2),
            "unique_recipients": len(profile["unique_recipients"]),
        }
    }

    return result


def store_verdict(req_id: str, result: dict):
    r.set(f"antifraud:verdict:{req_id}", json.dumps(result), ex=300)
    r.publish("antifraud:result", json.dumps(result))


def worker():
    print("[PY] Anti-fraud Python worker started")
    while True:
        try:
            _, raw = r.brpop("antifraud:queue:python", timeout=5)
            if not raw:
                continue
            req = json.loads(raw)
            print(f"[PY] <- {req.get('id', '?')[:16]}... amount={req.get('amount')}")
            result = analyze_transfer(req)
            store_verdict(req["id"], result)
            status = "APPROVED" if result["approved"] else "BLOCKED"
            print(f"[PY] -> {req.get('id', '?')[:16]}... {status} risk={result['risk_score']}")
        except Exception as e:
            print(f"[PY] Error: {e}")
            time.sleep(1)


def register_account_creation(account_id: str):
    r.set(f"account:created:{account_id}", str(time.time()), ex=86400)


if __name__ == "__main__":
    worker()
