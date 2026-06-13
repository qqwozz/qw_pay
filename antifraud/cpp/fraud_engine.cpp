#include <iostream>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <vector>
#include <chrono>
#include <mutex>
#include <sstream>
#include <algorithm>
#include <hiredis/hiredis.h>

struct TransferRequest {
    std::string id;
    std::string from_account;
    std::string to_account;
    double amount;
    std::string currency;
    std::string user_id;
};

struct FraudResult {
    bool approved;
    std::string reason;
    int risk_score;
};

class FraudEngine {
private:
    redisContext *rctx;
    std::mutex mtx;

    std::unordered_map<std::string, std::vector<long long>> transfer_times;
    std::unordered_map<std::string, double> daily_totals;
    std::unordered_set<std::string> blocked_users;
    std::unordered_set<std::string> blocked_accounts;
    std::unordered_map<std::string, std::unordered_set<std::string>> recipient_senders;

    const int MAX_TRANSFERS_PER_MINUTE = 5;
    const int MAX_TRANSFERS_PER_HOUR = 20;
    const double MAX_SINGLE_AMOUNT = 500000.0;
    const double DAILY_LIMIT = 2000000.0;
    const int MAX_UNIQUE_RECIPIENTS_PER_DAY = 10;

    long long now_ms() {
        return std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()
        ).count();
    }

    long long today_start_ms() {
        auto now = std::chrono::system_clock::now();
        auto t = std::chrono::system_clock::to_time_t(now);
        auto tm = *std::localtime(&t);
        tm.tm_hour = 0; tm.tm_min = 0; tm.tm_sec = 0;
        auto start = std::chrono::system_clock::from_time_t(std::mktime(&tm));
        return std::chrono::duration_cast<std::chrono::milliseconds>(
            start.time_since_epoch()
        ).count();
    }

    void cleanup_old_entries(const std::string& user_id) {
        long long one_hour_ago = now_ms() - 3600000;
        auto& times = transfer_times[user_id];
        times.erase(
            std::remove_if(times.begin(), times.end(),
                [one_hour_ago](long long t) { return t < one_hour_ago; }),
            times.end()
        );
    }

    void load_blacklists() {
        redisReply *reply;
        reply = (redisReply*)redisCommand(rctx, "SMEMBERS antifraud:blocked_users");
        if (reply && reply->type == REDIS_REPLY_ARRAY) {
            for (size_t i = 0; i < reply->elements; i++)
                blocked_users.insert(reply->element[i]->str);
        }
        freeReplyObject(reply);

        reply = (redisReply*)redisCommand(rctx, "SMEMBERS antifraud:blocked_accounts");
        if (reply && reply->type == REDIS_REPLY_ARRAY) {
            for (size_t i = 0; i < reply->elements; i++)
                blocked_accounts.insert(reply->element[i]->str);
        }
        freeReplyObject(reply);
    }

public:
    redisContext* getRedis() { return rctx; }

    FraudEngine(const char *host, int port) {
        rctx = redisConnect(host, port);
        if (rctx == NULL || rctx->err) {
            std::cerr << "[FATAL] Redis connection failed" << std::endl;
            exit(1);
        }
        load_blacklists();
        std::cout << "[CPP] FraudEngine connected to Redis " << host << ":" << port << std::endl;
    }

    ~FraudEngine() {
        if (rctx) redisFree(rctx);
    }

    FraudResult check(const TransferRequest& req) {
        std::lock_guard<std::mutex> lock(mtx);

        if (blocked_users.count(req.user_id))
            return {false, "User is blacklisted", 100};
        if (blocked_accounts.count(req.from_account) || blocked_accounts.count(req.to_account))
            return {false, "Account is blacklisted", 100};
        if (req.amount > MAX_SINGLE_AMOUNT)
            return {false, "Amount exceeds single transfer limit", 90};
        if (req.amount <= 0)
            return {false, "Invalid amount", 100};
        if (req.from_account == req.to_account)
            return {false, "Cannot transfer to same account", 80};

        cleanup_old_entries(req.user_id);
        auto& times = transfer_times[req.user_id];
        long long now = now_ms();

        int last_minute = 0;
        for (auto it = times.rbegin(); it != times.rend(); ++it) {
            if (now - *it <= 60000) last_minute++;
            else break;
        }
        if (last_minute >= MAX_TRANSFERS_PER_MINUTE)
            return {false, "Too many transfers per minute", 85};
        if ((int)times.size() >= MAX_TRANSFERS_PER_HOUR)
            return {false, "Hourly transfer limit reached", 80};

        long long day_start = today_start_ms();
        auto& daily = daily_totals[req.user_id];
        if (daily == 0) {
            redisReply *reply = (redisReply*)redisCommand(rctx,
                "GET antifraud:daily:%s", req.user_id.c_str());
            if (reply && reply->type == REDIS_REPLY_STRING)
                daily = std::stod(reply->str);
            freeReplyObject(reply);
        }
        if (daily + req.amount > DAILY_LIMIT)
            return {false, "Daily transfer limit exceeded", 75};

        auto& senders = recipient_senders[req.to_account];
        senders.insert(req.from_account);
        if ((int)senders.size() > MAX_UNIQUE_RECIPIENTS_PER_DAY)
            return {false, "Suspicious: too many unique senders to this account", 70};

        if (req.amount == (int)req.amount && req.amount >= 1000) {
            times.push_back(now);
            daily += req.amount;
            return {true, "Approved (flagged: round amount)", 30};
        }

        times.push_back(now);
        daily += req.amount;

        redisCommand(rctx, "SET antifraud:daily:%s %.2f EX 86400",
            req.user_id.c_str(), daily);

        int risk = 0;
        if (req.amount > 100000) risk = 20;
        else if (req.amount > 50000) risk = 10;

        return {true, "Approved", risk};
    }

    void block_user(const std::string& user_id) {
        std::lock_guard<std::mutex> lock(mtx);
        blocked_users.insert(user_id);
        redisCommand(rctx, "SADD antifraud:blocked_users %s", user_id.c_str());
    }

    void block_account(const std::string& account_id) {
        std::lock_guard<std::mutex> lock(mtx);
        blocked_accounts.insert(account_id);
        redisCommand(rctx, "SADD antifraud:blocked_accounts %s", account_id.c_str());
    }
};

TransferRequest parse_request(const std::string& s) {
    TransferRequest req;
    auto extract = [&](const std::string& key) -> std::string {
        auto pos = s.find("\"" + key + "\"");
        if (pos == std::string::npos) return "";
        pos = s.find(':', pos);
        if (pos == std::string::npos) return "";
        pos++;
        while (pos < s.size() && (s[pos] == ' ' || s[pos] == '"')) pos++;
        auto end = s.find_first_of(",}", pos);
        if (end == std::string::npos) end = s.size();
        return s.substr(pos, end - pos);
    };
    req.id = extract("id");
    req.from_account = extract("from_account");
    req.to_account = extract("to_account");
    req.user_id = extract("user_id");
    req.currency = extract("currency");
    std::string amt_str = extract("amount");
    try { req.amount = std::stod(amt_str); } catch(...) { req.amount = 0; }
    return req;
}

int main(int argc, char *argv[]) {
    const char *redis_host = "127.0.0.1";
    int redis_port = 6379;
    if (argc >= 2) redis_host = argv[1];
    if (argc >= 3) redis_port = std::atoi(argv[2]);

    FraudEngine engine(redis_host, redis_port);
    redisContext *rctx = engine.getRedis();
    std::cout << "[CPP] FraudEngine started, waiting on antifraud:queue" << std::endl;

    while (true) {
        redisReply *reply = (redisReply*)redisCommand(rctx, "BRPOP antifraud:queue 5");
        if (!reply) continue;

        if (reply->type == REDIS_REPLY_ARRAY && reply->elements >= 2) {
            std::string payload(reply->element[1]->str, reply->element[1]->len);
            std::cout << "[CPP] <- " << payload.substr(0, 150) << std::endl;

            TransferRequest req = parse_request(payload);
            FraudResult result = engine.check(req);

            std::ostringstream oss;
            oss << "{\"id\":\"" << req.id
                << "\",\"approved\":" << (result.approved ? "true" : "false")
                << ",\"reason\":\"" << result.reason
                << "\",\"risk_score\":" << result.risk_score
                << ",\"engine\":\"cpp\"}";

            std::string result_json = oss.str();
            redisCommand(rctx, "PUBLISH antifraud:result %s", result_json.c_str());
            redisCommand(rctx, "SET antifraud:verdict:%s %s EX 300",
                req.id.c_str(), result_json.c_str());

            std::cout << "[CPP] -> " << req.id << " : "
                      << (result.approved ? "APPROVED" : "BLOCKED")
                      << " risk=" << result.risk_score
                      << " [" << result.reason << "]" << std::endl;
        }
        freeReplyObject(reply);
    }
    return 0;
}
