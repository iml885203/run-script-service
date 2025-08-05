// API Client for Run Script Service
class APIClient {
    constructor(baseURL = '/api') {
        this.baseURL = baseURL;
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);
            const data = await response.json();

            if (!data.success) {
                throw new Error(data.error || 'API request failed');
            }

            return data.data;
        } catch (error) {
            console.error(`API request failed: ${endpoint}`, error);
            throw error;
        }
    }

    // System endpoints
    async getStatus() {
        return this.request('/status');
    }

    async getConfig() {
        return this.request('/config');
    }

    async updateConfig(config) {
        return this.request('/config', {
            method: 'PUT',
            body: JSON.stringify(config)
        });
    }

    // Script management endpoints
    async getScripts() {
        return this.request('/scripts');
    }

    async getScript(name) {
        return this.request(`/scripts/${encodeURIComponent(name)}`);
    }

    async addScript(script) {
        return this.request('/scripts', {
            method: 'POST',
            body: JSON.stringify(script)
        });
    }

    async updateScript(name, script) {
        return this.request(`/scripts/${encodeURIComponent(name)}`, {
            method: 'PUT',
            body: JSON.stringify(script)
        });
    }

    async deleteScript(name) {
        return this.request(`/scripts/${encodeURIComponent(name)}`, {
            method: 'DELETE'
        });
    }

    async runScript(name) {
        return this.request(`/scripts/${encodeURIComponent(name)}/run`, {
            method: 'POST'
        });
    }

    async enableScript(name) {
        return this.request(`/scripts/${encodeURIComponent(name)}/enable`, {
            method: 'POST'
        });
    }

    async disableScript(name) {
        return this.request(`/scripts/${encodeURIComponent(name)}/disable`, {
            method: 'POST'
        });
    }

    // Log management endpoints
    async getLogs(options = {}) {
        const params = new URLSearchParams();

        if (options.script) {
            params.append('script', options.script);
        }
        if (options.limit) {
            params.append('limit', options.limit);
        }

        const query = params.toString();
        const endpoint = query ? `/logs?${query}` : '/logs';

        return this.request(endpoint);
    }

    async getScriptLogs(scriptName, options = {}) {
        const params = new URLSearchParams();

        if (options.limit) {
            params.append('limit', options.limit);
        }

        const query = params.toString();
        const endpoint = query ?
            `/logs/${encodeURIComponent(scriptName)}?${query}` :
            `/logs/${encodeURIComponent(scriptName)}`;

        return this.request(endpoint);
    }

    async clearScriptLogs(scriptName) {
        return this.request(`/logs/${encodeURIComponent(scriptName)}`, {
            method: 'DELETE'
        });
    }
}

// Create global API client instance
window.apiClient = new APIClient();
