// UI Components and Utilities
class ComponentManager {
    constructor() {
        this.notifications = [];
    }

    // Notification system
    showNotification(message, type = 'info', title = '', duration = 5000) {
        const notification = this.createNotification(message, type, title);
        const container = document.getElementById('notification-container');

        if (container) {
            container.appendChild(notification);
            this.notifications.push(notification);

            // Auto-remove after duration
            if (duration > 0) {
                setTimeout(() => {
                    this.removeNotification(notification);
                }, duration);
            }
        }
    }

    createNotification(message, type, title) {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;

        notification.innerHTML = `
            <button class="notification-close">&times;</button>
            ${title ? `<div class="notification-title">${this.escapeHtml(title)}</div>` : ''}
            <div class="notification-message">${this.escapeHtml(message)}</div>
        `;

        // Add close handler
        const closeBtn = notification.querySelector('.notification-close');
        closeBtn.addEventListener('click', () => {
            this.removeNotification(notification);
        });

        return notification;
    }

    removeNotification(notification) {
        const index = this.notifications.indexOf(notification);
        if (index > -1) {
            this.notifications.splice(index, 1);
        }

        if (notification.parentNode) {
            notification.style.animation = 'slideOut 0.3s ease-in';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }
    }

    // Script table rendering
    renderScriptTable(scripts) {
        const tbody = document.getElementById('scripts-tbody');
        if (!tbody) return;

        if (!scripts || scripts.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <div class="empty-state-icon">📝</div>
                        <div class="empty-state-title">No Scripts</div>
                        <div class="empty-state-message">Add your first script to get started</div>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = scripts.map(script => `
            <tr>
                <td>
                    <strong>${this.escapeHtml(script.name)}</strong>
                    <br>
                    <small class="text-secondary">${this.escapeHtml(script.path)}</small>
                </td>
                <td>
                    <span class="status-badge ${script.enabled ? 'status-enabled' : 'status-disabled'}">
                        ${script.enabled ? '✓ Enabled' : '○ Disabled'}
                    </span>
                    ${script.running ? '<span class="status-badge status-running">Running</span>' : ''}
                </td>
                <td>
                    <span class="text-secondary">-</span>
                </td>
                <td>
                    <span>${script.interval}s</span>
                </td>
                <td>
                    <div class="action-buttons">
                        <button class="btn btn-sm btn-primary" onclick="app.runScript('${this.escapeHtml(script.name)}')">
                            ▶ Run
                        </button>
                        <button class="btn btn-sm ${script.enabled ? 'btn-secondary' : 'btn-primary'}"
                                onclick="app.${script.enabled ? 'disableScript' : 'enableScript'}('${this.escapeHtml(script.name)}')">
                            ${script.enabled ? '⏸ Disable' : '▶ Enable'}
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="app.deleteScript('${this.escapeHtml(script.name)}')">
                            🗑 Delete
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');
    }

    // Log rendering (simplified for raw text content)
    renderLogs(logData) {
        const logsList = document.getElementById('logs-list');
        if (!logsList) return;

        // Handle new API format: {content: "...", script: "..."}
        if (!logData) {
            logsList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📄</div>
                    <div class="empty-state-title">No Logs</div>
                    <div class="empty-state-message">Select a script to see logs here</div>
                </div>
            `;
            return;
        }

        // Check if no script is selected
        if (!logData.script || logData.script === '') {
            logsList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📄</div>
                    <div class="empty-state-title">No Script Selected</div>
                    <div class="empty-state-message">Please select a script from the dropdown to view logs</div>
                </div>
            `;
            return;
        }

        const content = logData.content ? logData.content.trim() : '';
        if (!content) {
            logsList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📄</div>
                    <div class="empty-state-title">No Logs</div>
                    <div class="empty-state-message">No logs found for ${this.escapeHtml(logData.script)}</div>
                </div>
            `;
            return;
        }

        // Display raw log content in a pre-formatted container
        logsList.innerHTML = `
            <div class="log-content">
                <div class="log-content-header">
                    <span class="log-content-script">Logs for: ${this.escapeHtml(logData.script)}</span>
                    <button class="btn btn-sm btn-secondary" onclick="app.clearSpecificScriptLogs('${this.escapeHtml(logData.script)}')">
                        Clear Logs
                    </button>
                </div>
                <pre class="log-content-text">${this.escapeHtml(content)}</pre>
            </div>
        `;
    }

    // Script filter dropdown
    updateScriptFilter(scripts) {
        const filter = document.getElementById('script-filter');
        if (!filter) return;

        // Keep current selection
        const currentValue = filter.value;

        // Start with empty option prompting user to select
        filter.innerHTML = '<option value="">Select a script...</option>';

        if (scripts && scripts.length > 0) {
            scripts.forEach(script => {
                const option = document.createElement('option');
                option.value = script.name;
                option.textContent = script.name;
                option.selected = script.name === currentValue;
                filter.appendChild(option);
            });
        }
    }

    // Dashboard metrics
    updateDashboardMetrics(scripts, systemStatus) {
        // Update system status
        const statusElement = document.getElementById('system-status');
        if (statusElement) {
            const statusDot = statusElement.querySelector('.status-dot');
            const statusText = statusElement.querySelector('span:last-child');

            if (systemStatus && systemStatus.status === 'running') {
                statusDot.className = 'status-dot status-running';
                if (statusText) statusText.textContent = 'Running';
            } else {
                statusDot.className = 'status-dot status-error';
                if (statusText) statusText.textContent = 'Error';
            }
        }

        // Update active scripts count
        const activeScriptsElement = document.getElementById('active-scripts-count');
        if (activeScriptsElement && scripts) {
            const activeCount = scripts.filter(script => script.enabled).length;
            const valueElement = activeScriptsElement.querySelector('.metric-value');
            if (valueElement) {
                valueElement.textContent = activeCount;
            }
        }
    }

    // Modal handling
    openModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('active');
            document.body.style.overflow = 'hidden';
        }
    }

    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('active');
            document.body.style.overflow = '';
        }
    }

    // Form handling
    getFormData(formId) {
        const form = document.getElementById(formId);
        if (!form) return null;

        const formData = new FormData(form);
        const data = {};

        for (const [key, value] of formData.entries()) {
            if (form.elements[key].type === 'checkbox') {
                data[key] = form.elements[key].checked;
            } else if (form.elements[key].type === 'number') {
                data[key] = parseInt(value, 10) || 0;
            } else {
                data[key] = value;
            }
        }

        return data;
    }

    resetForm(formId) {
        const form = document.getElementById(formId);
        if (form) {
            form.reset();
        }
    }

    // Utility functions
    escapeHtml(text) {
        if (typeof text !== 'string') return text;
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    formatDuration(milliseconds) {
        if (milliseconds < 1000) return `${milliseconds}ms`;
        const seconds = Math.floor(milliseconds / 1000);
        if (seconds < 60) return `${seconds}s`;
        const minutes = Math.floor(seconds / 60);
        return `${minutes}m ${seconds % 60}s`;
    }

    // Loading states
    showLoading(elementId, message = 'Loading...') {
        const element = document.getElementById(elementId);
        if (element) {
            element.innerHTML = `
                <div class="loading">
                    <span class="spinner"></span>
                    ${message}
                </div>
            `;
        }
    }

    hideLoading(elementId) {
        const element = document.getElementById(elementId);
        if (element) {
            const loading = element.querySelector('.loading');
            if (loading) {
                loading.remove();
            }
        }
    }
}

// Create global components manager
window.components = new ComponentManager();

// Add CSS for slideOut animation
const style = document.createElement('style');
style.textContent = `
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);
