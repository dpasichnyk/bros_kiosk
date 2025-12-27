class ClockManager {
    constructor(config) {
        this.config = config;
        this.timeEl = document.getElementById('clock-time');
        this.dateEl = document.getElementById('clock-date');
        this.timeFormatter = new Intl.DateTimeFormat(config.locale, {
            hour: 'numeric',
            minute: '2-digit',
            hour12: config.timeFormat === '12h'
        });
        this.dateFormatter = new Intl.DateTimeFormat(config.locale, {
            weekday: 'long',
            month: 'long',
            day: 'numeric'
        });
        this.start();
    }

    start() {
        this.update();
        setInterval(() => this.update(), 1000);
    }

    update() {
        const now = new Date();
        this.timeEl.textContent = this.timeFormatter.format(now);
        this.dateEl.textContent = this.dateFormatter.format(now);
    }
}

class SlideshowManager {
    constructor(config) {
        this.config = config;
        this.container = document.getElementById('slideshow');
        this.slides = Array.from(this.container.querySelectorAll('.slide'));
        this.currentIndex = 0;
        this.photos = [];
        this.init();
    }

    async init() {
        try {
            const resp = await fetch('/api/photos');
            const data = await resp.json();
            this.photos = data.photos;
            if (this.photos.length > 0) {
                this.start();
            }
        } catch (e) {
            console.error("Failed to fetch photos:", e);
        }
    }

    start() {
        this.loadImage(this.slides[0], this.photos[0]);
        const intervalSec = parseInt(this.config.slideshow.interval) || 30;
        setInterval(() => this.nextSlide(), intervalSec * 1000);
    }

    nextSlide() {
        const nextIndex = (this.currentIndex + 1) % this.photos.length;
        const currentSlide = this.slides[0];
        const nextSlide = this.slides[1];

        // Load new image into next slide
        this.loadImage(nextSlide, this.photos[nextIndex]).then(() => {
            nextSlide.classList.add('active');
            currentSlide.classList.remove('active');

            // Wait for transition to finish, then clean up
            setTimeout(() => {
                currentSlide.style.backgroundImage = 'none';

                // Revoke the old object URL if it exists to free memory
                if (currentSlide._objectUrl) {
                    URL.revokeObjectURL(currentSlide._objectUrl);
                    currentSlide._objectUrl = null;
                }
            }, 2000); // slightly longer than CSS transition

            this.slides.reverse();
            this.currentIndex = nextIndex;
        });
    }

    async loadImage(el, path) {
        try {
            const encodedPath = encodeURIComponent(path);
            const resp = await fetch(`/assets/photos/${encodedPath}`);
            if (!resp.ok) throw new Error('Failed to load image');

            const blob = await resp.blob();
            const objectUrl = URL.createObjectURL(blob);

            // Clean up any existing URL on this element before assigning new one
            if (el._objectUrl) {
                URL.revokeObjectURL(el._objectUrl);
            }

            el.style.backgroundImage = `url('${objectUrl}')`;
            el._objectUrl = objectUrl;
        } catch (e) {
            console.error("Error loading slide:", e);
        }
    }
}

class DashboardClient {
    constructor(config) {
        this.config = config;
        this.hash = "";
        this.dateFormatter = new Intl.DateTimeFormat(config.locale, {
            month: 'short', day: 'numeric'
        });
        this.timeFormatter = new Intl.DateTimeFormat(config.locale, {
            hour: 'numeric', minute: '2-digit', hour12: config.timeFormat === '12h'
        });
        this.relativeTimeFormatter = new Intl.DateTimeFormat(config.locale, {
            hour: 'numeric', minute: '2-digit'
        });
        this.connect();
    }

    async connect() {
        const backoffBase = 1000;
        let attempts = 0;

        while (true) {
            try {
                await this.poll();
                attempts = 0;
            } catch (e) {
                console.error("Poll failed:", e);
                attempts++;
                const delay = Math.min(backoffBase * Math.pow(2, attempts), 30000);
                await new Promise(r => setTimeout(r, delay));
            }
        }
    }

    async poll() {
        const headers = {};
        if (this.hash) {
            headers['X-Dashboard-Hash'] = this.hash;
        }

        const resp = await fetch('/api/updates', { headers });

        if (resp.status === 304) {
            return;
        }

        if (resp.ok) {
            const data = await resp.json();
            this.hash = data.hash;
            this.updateDOM(data.updates);
        } else {
            throw new Error(`Server returned ${resp.status}`);
        }
    }

    updateDOM(updates) {
        for (const [id, result] of Object.entries(updates)) {
            const section = document.getElementById(`section-${id}`);
            if (!section) continue;

            const type = section.dataset.type;

            if (result.status && result.status.error) {
                console.warn(`Error for ${id}:`, result.status.error);
                continue;
            }

            if (result.data) {
                this.renderSection(section, type, result.data);
            }
        }
    }

    renderSection(el, type, data) {
        switch (type) {
            case 'weather':
                this.renderWeather(el, data);
                break;
            case 'rss':
                this.renderNews(el, data);
                break;
            case 'calendar':
                this.renderCalendar(el, data);
                break;
        }
    }

    mapWeatherIcon(iconCode) {
        if (!iconCode) return 'wi-default';

        const code = iconCode.replace('n', 'd');
        const isNight = iconCode.includes('n');

        const iconMap = {
            '01d': isNight ? 'wi-night' : 'wi-sunny',
            '02d': 'wi-partly-cloudy',
            '03d': 'wi-cloudy',
            '04d': 'wi-cloudy',
            '09d': 'wi-rain',
            '10d': 'wi-rain',
            '11d': 'wi-thunderstorm',
            '13d': 'wi-snow',
            '50d': 'wi-fog'
        };

        return iconMap[code] || 'wi-default';
    }

    renderWeather(el, data) {
        if (data.setup_required) {
            el.querySelector('.module-content').innerHTML = '<div class="loading">Setup Required</div>';
            return;
        }

        const tempEl = el.querySelector('[data-field="temp"]');
        const condEl = el.querySelector('[data-field="condition"]');
        const cityEl = el.querySelector('[data-field="city"]');
        const iconEl = el.querySelector('[data-field="icon"]');

        if (tempEl) tempEl.textContent = `${Math.round(data.temp)}°`;
        if (condEl) condEl.textContent = data.description;
        if (cityEl) cityEl.textContent = data.city;
        if (iconEl) {
            const iconClass = this.mapWeatherIcon(data.icon);
            iconEl.className = `weather-icon wi ${iconClass}`;
        }
    }

    renderNews(el, data) {
        const container = el.querySelector('[data-field="items"]');
        if (!data.items || data.items.length === 0) {
            container.innerHTML = '<div class="loading">No news</div>';
            return;
        }

        container.innerHTML = data.items.map(item => {
            // Summary is already cleaned by the server
            let summary = item.summary || '';
            const title = item.title || '';

            return `
            <div class="news-item">
                <div class="news-title">${this.escapeHtml(title)}</div>
                ${summary ? `<div class="news-summary">${this.escapeHtml(summary)}</div>` : ''}
                <div class="news-time">${this.formatRelativeTime(item.pub_date)}</div>
            </div>
        `}).join('');
    }


    renderCalendar(el, data) {
        const container = el.querySelector('[data-field="events"]');
        if (!data.events || data.events.length === 0) {
            container.innerHTML = '<div class="loading">No upcoming events</div>';
            return;
        }

        container.innerHTML = data.events.map(event => {
            const startTime = new Date(event.start);
            const dateStr = this.dateFormatter.format(startTime);
            const timeStr = event.all_day ? 'All Day' : this.timeFormatter.format(startTime);

            return `
            <div class="event-item">
                <div class="event-time-badge">
                    <div class="event-date">${dateStr}</div>
                    <div class="event-time">${timeStr}</div>
                </div>
                <div class="event-details">
                    <div class="event-title">${this.escapeHtml(event.summary)}</div>
                    ${event.location ? `<div class="event-location">${this.escapeHtml(event.location)}</div>` : ''}
                </div>
            </div>
        `}).join('');
    }

    formatRelativeTime(dateStr) {
        const date = new Date(dateStr);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMins / 60);

        if (diffMins < 60) {
            return `${diffMins}m ago`;
        } else if (diffHours < 24) {
            return `${diffHours}h ago`;
        }
        return this.relativeTimeFormatter.format(date);
    }

    escapeHtml(text) {
        if (!text) return '';
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }


}

document.addEventListener('DOMContentLoaded', () => {
    const config = window.KIOSK_CONFIG;
    new ClockManager(config);
    new SlideshowManager(config);
    new DashboardClient(config);
});
