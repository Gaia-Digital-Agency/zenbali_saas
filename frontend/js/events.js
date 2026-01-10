/* ===========================================
   Zen Bali - Events JavaScript
   Developed by net1io.com
   Copyright (C) 2024
   =========================================== */

let currentPage = 1;
let totalPages = 1;
let locations = [];
let eventTypes = [];
let entranceTypes = [];

// Load filter options
async function loadFilters() {
    try {
        // Load locations
        const locResponse = await API.get('/locations');
        if (locResponse.success) {
            locations = locResponse.data;
            populateSelect('filterLocation', locations);
        }

        // Load event types
        const typeResponse = await API.get('/event-types');
        if (typeResponse.success) {
            eventTypes = typeResponse.data;
            populateSelect('filterEventType', eventTypes);
        }

        // Load entrance types
        const entranceResponse = await API.get('/entrance-types');
        if (entranceResponse.success) {
            entranceTypes = entranceResponse.data;
            populateSelect('filterEntranceType', entranceTypes);
        }

        // Setup filter form
        setupFilterForm();

        // Load from URL params
        loadFiltersFromURL();
    } catch (error) {
        console.error('Failed to load filters:', error);
    }
}

function populateSelect(selectId, items) {
    const select = document.getElementById(selectId);
    if (!select) return;

    items.forEach(item => {
        const option = document.createElement('option');
        option.value = item.id;
        option.textContent = item.name;
        select.appendChild(option);
    });
}

function setupFilterForm() {
    const form = document.getElementById('filterForm');
    if (!form) return;

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        currentPage = 1;
        loadEvents();
        saveFiltersToURL();
    });

    // Auto-submit on select change
    form.querySelectorAll('select').forEach(select => {
        select.addEventListener('change', () => {
            currentPage = 1;
            loadEvents();
            saveFiltersToURL();
        });
    });

    // Debounced search
    let searchTimeout;
    const searchInput = document.getElementById('filterSearch');
    if (searchInput) {
        searchInput.addEventListener('input', () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                currentPage = 1;
                loadEvents();
                saveFiltersToURL();
            }, 500);
        });
    }
}

function loadFiltersFromURL() {
    const params = new URLSearchParams(window.location.search);
    
    const locationId = params.get('location_id');
    if (locationId) {
        const select = document.getElementById('filterLocation');
        if (select) select.value = locationId;
    }

    const eventTypeId = params.get('event_type_id');
    if (eventTypeId) {
        const select = document.getElementById('filterEventType');
        if (select) select.value = eventTypeId;
    }

    const entranceTypeId = params.get('entrance_type_id');
    if (entranceTypeId) {
        const select = document.getElementById('filterEntranceType');
        if (select) select.value = entranceTypeId;
    }

    const dateFrom = params.get('date_from');
    if (dateFrom) {
        const input = document.getElementById('filterDateFrom');
        if (input) input.value = dateFrom;
    }

    const search = params.get('search');
    if (search) {
        const input = document.getElementById('filterSearch');
        if (input) input.value = search;
    }

    const page = params.get('page');
    if (page) {
        currentPage = parseInt(page, 10) || 1;
    }
}

function saveFiltersToURL() {
    const params = {};
    
    const locationId = document.getElementById('filterLocation')?.value;
    if (locationId) params.location_id = locationId;

    const eventTypeId = document.getElementById('filterEventType')?.value;
    if (eventTypeId) params.event_type_id = eventTypeId;

    const entranceTypeId = document.getElementById('filterEntranceType')?.value;
    if (entranceTypeId) params.entrance_type_id = entranceTypeId;

    const dateFrom = document.getElementById('filterDateFrom')?.value;
    if (dateFrom) params.date_from = dateFrom;

    const search = document.getElementById('filterSearch')?.value;
    if (search) params.search = search;

    if (currentPage > 1) params.page = currentPage;

    Utils.setQueryParams(params);
}

async function loadEvents() {
    const container = document.getElementById('eventsContainer');
    if (!container) return;

    container.innerHTML = '<div class="loading loading-lg"><div class="spinner"></div></div>';

    try {
        // Build query string
        const params = new URLSearchParams();
        params.append('page', currentPage);
        params.append('limit', 12);

        const locationId = document.getElementById('filterLocation')?.value;
        if (locationId) params.append('location_id', locationId);

        const eventTypeId = document.getElementById('filterEventType')?.value;
        if (eventTypeId) params.append('event_type_id', eventTypeId);

        const entranceTypeId = document.getElementById('filterEntranceType')?.value;
        if (entranceTypeId) params.append('entrance_type_id', entranceTypeId);

        const dateFrom = document.getElementById('filterDateFrom')?.value;
        if (dateFrom) params.append('date_from', dateFrom);

        const search = document.getElementById('filterSearch')?.value;
        if (search) params.append('search', search);

        const response = await API.get(`/events?${params.toString()}`);

        if (response.success) {
            const { events, total, page, limit, total_pages } = response.data;
            totalPages = total_pages;
            currentPage = page;

            if (!events || events.length === 0) {
                container.innerHTML = `
                    <div class="empty-state">
                        <div class="empty-state-icon">🎭</div>
                        <h3 class="empty-state-title">No events found</h3>
                        <p class="empty-state-text">Try adjusting your filters or check back later for new events.</p>
                        <a href="/creator/register.html" class="btn btn-primary">Post an Event</a>
                    </div>
                `;
            } else {
                container.innerHTML = `<div class="event-grid">${events.map(renderEventCard).join('')}</div>`;
            }

            renderPagination();
        }
    } catch (error) {
        container.innerHTML = `
            <div class="alert alert-error">
                Failed to load events. Please try again later.
            </div>
        `;
        console.error('Failed to load events:', error);
    }
}

function renderEventCard(event) {
    const imageUrl = event.image_url || '/assets/images/placeholder.jpg';
    const eventDate = Utils.formatDate(event.event_date);
    const price = event.entrance_fee > 0 
        ? Utils.formatCurrency(event.entrance_fee) 
        : 'Free';

    return `
        <div class="event-card">
            <div class="event-card-image">
                <img src="${Utils.escapeHtml(imageUrl)}" alt="${Utils.escapeHtml(event.title)}" 
                     onerror="this.src='/assets/images/placeholder.jpg'">
                <span class="event-card-badge">${Utils.escapeHtml(event.event_type)}</span>
            </div>
            <div class="event-card-content">
                <div class="event-card-date">
                    <span>📅</span>
                    <span>${eventDate}${event.event_time ? ' • ' + event.event_time : ''}</span>
                </div>
                <h3 class="event-card-title">
                    <a href="/event.html?id=${event.id}">${Utils.escapeHtml(event.title)}</a>
                </h3>
                <div class="event-card-location">
                    <span>📍</span>
                    <span>${Utils.escapeHtml(event.location)}</span>
                </div>
                <p class="event-card-organizer">
                    by ${Utils.escapeHtml(event.organization_name || event.organizer)}
                </p>
            </div>
            <div class="event-card-footer">
                <span class="event-card-price">${price}</span>
                <span class="event-card-type">${Utils.escapeHtml(event.entrance_type)}</span>
            </div>
        </div>
    `;
}

function renderPagination() {
    const container = document.getElementById('pagination');
    if (!container || totalPages <= 1) {
        if (container) container.innerHTML = '';
        return;
    }

    let html = '';

    // Previous button
    html += `
        <button class="pagination-btn" onclick="goToPage(${currentPage - 1})" 
                ${currentPage === 1 ? 'disabled' : ''}>
            ← Prev
        </button>
    `;

    // Page numbers
    const startPage = Math.max(1, currentPage - 2);
    const endPage = Math.min(totalPages, startPage + 4);

    if (startPage > 1) {
        html += `<button class="pagination-btn" onclick="goToPage(1)">1</button>`;
        if (startPage > 2) {
            html += `<span style="padding: 0 0.5rem;">...</span>`;
        }
    }

    for (let i = startPage; i <= endPage; i++) {
        html += `
            <button class="pagination-btn ${i === currentPage ? 'active' : ''}" 
                    onclick="goToPage(${i})">
                ${i}
            </button>
        `;
    }

    if (endPage < totalPages) {
        if (endPage < totalPages - 1) {
            html += `<span style="padding: 0 0.5rem;">...</span>`;
        }
        html += `<button class="pagination-btn" onclick="goToPage(${totalPages})">${totalPages}</button>`;
    }

    // Next button
    html += `
        <button class="pagination-btn" onclick="goToPage(${currentPage + 1})" 
                ${currentPage === totalPages ? 'disabled' : ''}>
            Next →
        </button>
    `;

    container.innerHTML = html;
}

function goToPage(page) {
    if (page < 1 || page > totalPages) return;
    currentPage = page;
    loadEvents();
    saveFiltersToURL();
    window.scrollTo({ top: 0, behavior: 'smooth' });
}
