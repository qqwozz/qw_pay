# Frontend Architect

**ID:** frontend-architect
**Version:** 2.0
**Category:** Frontend & UI/UX
**Triggers:** frontend, UI, UX, React, Vue, CSS, HTML, responsive, accessibility, design system

---

## Role

I am a senior frontend architect. I design and implement user interfaces following modern web standards, accessibility guidelines, and performance best practices.

---

## Tech Stack

- **HTML5** — Semantic markup
- **CSS3** — Custom properties, Grid, Flexbox
- **JavaScript** — Vanilla ES6+ (no framework for demo)
- **Fetch API** — HTTP requests

---

## Project Structure

```
web/
├── index.html              # Main demo page
├── css/
│   ├── variables.css       # CSS custom properties
│   ├── base.css            # Reset and typography
│   ├── components.css      # UI components
│   └── responsive.css      # Media queries
├── js/
│   ├── api.js              # API client
│   ├── auth.js             # Authentication logic
│   ├── accounts.js         # Account management
│   ├── transactions.js     # Transfer logic
│   └── utils.js            # Utility functions
└── assets/
    ├── icons/              # SVG icons
    └── images/             # Images
```

---

## Design System

### CSS Variables
```css
:root {
    /* Colors */
    --color-primary: #3b82f6;
    --color-primary-hover: #2563eb;
    --color-success: #22c55e;
    --color-warning: #f59e0b;
    --color-danger: #ef4444;
    --color-bg: #ffffff;
    --color-bg-secondary: #f8fafc;
    --color-text: #1e293b;
    --color-text-secondary: #64748b;
    --color-border: #e2e8f0;
    
    /* Typography */
    --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    --font-mono: 'SF Mono', 'Fira Code', monospace;
    --font-size-xs: 0.75rem;
    --font-size-sm: 0.875rem;
    --font-size-base: 1rem;
    --font-size-lg: 1.125rem;
    --font-size-xl: 1.25rem;
    --font-size-2xl: 1.5rem;
    
    /* Spacing */
    --space-1: 0.25rem;
    --space-2: 0.5rem;
    --space-3: 0.75rem;
    --space-4: 1rem;
    --space-6: 1.5rem;
    --space-8: 2rem;
    
    /* Borders */
    --radius-sm: 0.25rem;
    --radius-md: 0.375rem;
    --radius-lg: 0.5rem;
    --radius-full: 9999px;
    
    /* Shadows */
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
    
    /* Transitions */
    --transition-fast: 150ms ease;
    --transition-base: 200ms ease;
    --transition-slow: 300ms ease;
}
```

### Component Styles

#### Buttons
```css
.btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-2) var(--space-4);
    font-size: var(--font-size-sm);
    font-weight: 500;
    border-radius: var(--radius-md);
    transition: all var(--transition-fast);
    cursor: pointer;
    border: none;
}

.btn-primary {
    background-color: var(--color-primary);
    color: white;
}

.btn-primary:hover {
    background-color: var(--color-primary-hover);
}

.btn-danger {
    background-color: var(--color-danger);
    color: white;
}

.btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}
```

#### Forms
```css
.form-group {
    margin-bottom: var(--space-4);
}

.form-label {
    display: block;
    margin-bottom: var(--space-1);
    font-size: var(--font-size-sm);
    font-weight: 500;
    color: var(--color-text);
}

.form-input {
    width: 100%;
    padding: var(--space-2) var(--space-3);
    font-size: var(--font-size-base);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    transition: border-color var(--transition-fast);
}

.form-input:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px rgb(59 130 246 / 0.1);
}

.form-input.error {
    border-color: var(--color-danger);
}
```

#### Cards
```css
.card {
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    padding: var(--space-6);
    box-shadow: var(--shadow-sm);
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-4);
}

.card-title {
    font-size: var(--font-size-lg);
    font-weight: 600;
    color: var(--color-text);
}
```

---

## API Client

```javascript
const API_BASE = '/api/v1';

class ApiClient {
    constructor() {
        this.token = localStorage.getItem('access_token');
    }

    async request(path, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers,
        };

        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }

        const response = await fetch(`${API_BASE}${path}`, {
            ...options,
            headers,
        });

        if (response.status === 401) {
            const refreshed = await this.refreshToken();
            if (refreshed) {
                headers['Authorization'] = `Bearer ${this.token}`;
                return fetch(`${API_BASE}${path}`, { ...options, headers });
            }
            this.logout();
            throw new Error('Session expired');
        }

        return response.json();
    }

    async get(path) {
        return this.request(path, { method: 'GET' });
    }

    async post(path, body) {
        return this.request(path, {
            method: 'POST',
            body: JSON.stringify(body),
        });
    }

    async refreshToken() {
        const refreshToken = localStorage.getItem('refresh_token');
        if (!refreshToken) return false;

        try {
            const response = await fetch(`${API_BASE}/refresh`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ refresh_token: refreshToken }),
            });

            if (response.ok) {
                const data = await response.json();
                this.token = data.data.access_token;
                localStorage.setItem('access_token', this.token);
                localStorage.setItem('refresh_token', data.data.refresh_token);
                return true;
            }
        } catch (e) {
            console.error('Refresh failed:', e);
        }
        return false;
    }

    logout() {
        this.token = null;
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        window.location.href = '/';
    }
}

const api = new ApiClient();
```

---

## Accessibility (WCAG 2.1 AA)

### Semantic HTML
```html
<!-- Use semantic elements -->
<header role="banner">
<nav role="navigation" aria-label="Main">
<main role="main">
<aside role="complementary">
<footer role="contentinfo">

<!-- Forms -->
<label for="email">Email</label>
<input type="email" id="email" aria-required="true" aria-invalid="false">
<span id="email-error" role="alert" aria-live="polite"></span>

<!-- Buttons -->
<button type="submit" aria-label="Submit transfer">
    <span aria-hidden="true">→</span> Send
</button>

<!-- Tables -->
<table aria-label="Transaction history">
    <thead>
        <tr>
            <th scope="col">Date</th>
            <th scope="col">Amount</th>
        </tr>
    </thead>
</table>
```

### Keyboard Navigation
```javascript
// Trap focus in modal
function trapFocus(element) {
    const focusableElements = element.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    element.addEventListener('keydown', (e) => {
        if (e.key !== 'Tab') return;

        if (e.shiftKey) {
            if (document.activeElement === firstElement) {
                lastElement.focus();
                e.preventDefault();
            }
        } else {
            if (document.activeElement === lastElement) {
                firstElement.focus();
                e.preventDefault();
            }
        }
    });
}
```

---

## Performance Optimization

### Lazy Loading
```html
<!-- Images -->
<img src="placeholder.svg" data-src="actual-image.jpg" loading="lazy" alt="Description">

<!-- Intersection Observer -->
<script>
const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            const img = entry.target;
            img.src = img.dataset.src;
            observer.unobserve(img);
        }
    });
});

document.querySelectorAll('img[data-src]').forEach(img => observer.observe(img));
</script>
```

### Debounce
```javascript
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Usage
searchInput.addEventListener('input', debounce((e) => {
    api.get(`/transactions?search=${e.target.value}`);
}, 300));
```

---

## Responsive Design

```css
/* Mobile first */
.container {
    padding: var(--space-4);
}

/* Tablet */
@media (min-width: 768px) {
    .container {
        padding: var(--space-6);
        max-width: 720px;
        margin: 0 auto;
    }
}

/* Desktop */
@media (min-width: 1024px) {
    .container {
        max-width: 960px;
    }
}

/* Grid layout */
.grid {
    display: grid;
    gap: var(--space-4);
}

@media (min-width: 768px) {
    .grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (min-width: 1024px) {
    .grid {
        grid-template-columns: repeat(3, 1fr);
    }
}
```

---

## Implementation Checklist

- [ ] Semantic HTML structure
- [ ] CSS variables for theming
- [ ] Responsive design (mobile/tablet/desktop)
- [ ] Keyboard navigation
- [ ] ARIA labels
- [ ] Error handling with user feedback
- [ ] Loading states
- [ ] API error handling
- [ ] Token refresh logic
- [ ] Form validation
- [ ] XSS prevention (sanitize inputs)
- [ ] Performance optimization
