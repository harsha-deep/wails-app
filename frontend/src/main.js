import './style.css';
import './app.css';

import { GetSystemStats } from '../wailsjs/go/main/App';

let updateInterval = null;
let lastUpdate = Date.now();

// Format bytes to human readable format
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// Format uptime to human readable format
function formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    const parts = [];
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);

    return parts.length > 0 ? parts.join(' ') : '0m';
}

// Get process state class
function getProcessStateClass(state) {
    if (state === 'R') return 'running';
    if (state === 'S' || state === 'D') return 'sleeping';
    if (state === 'T' || state === 'Z') return 'stopped';
    return 'sleeping';
}

// Get process state name
function getProcessStateName(state) {
    const states = {
        'R': 'Running',
        'S': 'Sleeping',
        'D': 'Disk Sleep',
        'T': 'Stopped',
        'Z': 'Zombie',
        'X': 'Dead'
    };
    return states[state] || state;
}

// Create the UI
function createUI() {
    document.querySelector('#app').innerHTML = `
        <div class="header">
            <h1>System Monitor</h1>
            <div class="header-stats">
                <div class="header-stat">
                    <div class="header-stat-label">CPU Usage</div>
                    <div class="header-stat-value" id="header-cpu">--</div>
                </div>
                <div class="header-stat">
                    <div class="header-stat-label">Memory Usage</div>
                    <div class="header-stat-value" id="header-memory">--</div>
                </div>
                <div class="header-stat">
                    <div class="header-stat-label">Processes</div>
                    <div class="header-stat-value" id="header-processes">--</div>
                </div>
                <div class="header-stat">
                    <div class="header-stat-label">Uptime</div>
                    <div class="header-stat-value" id="header-uptime">--</div>
                </div>
            </div>
        </div>

        <div class="content">
            <div class="section">
                <div class="section-title">
                    <span class="cpu-icon">🖥️</span>
                    CPU
                </div>
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-label">Usage</div>
                        <div class="stat-value" id="cpu-usage">--</div>
                        <div class="progress-bar">
                            <div class="progress-fill" id="cpu-progress" style="width: 0%"></div>
                        </div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Cores</div>
                        <div class="stat-value" id="cpu-cores">--</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Model</div>
                        <div class="stat-value" style="font-size: 14px;" id="cpu-model">--</div>
                    </div>
                </div>
            </div>

            <div class="section">
                <div class="section-title">
                    <span class="memory-icon">💾</span>
                    Memory
                </div>
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-label">Used</div>
                        <div class="stat-value" id="memory-used">--</div>
                        <div class="stat-subvalue" id="memory-used-percent">--</div>
                        <div class="progress-bar">
                            <div class="progress-fill" id="memory-progress" style="width: 0%"></div>
                        </div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Total</div>
                        <div class="stat-value" id="memory-total">--</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Available</div>
                        <div class="stat-value" id="memory-available">--</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Cached</div>
                        <div class="stat-value" id="memory-cached">--</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Swap Used</div>
                        <div class="stat-value" id="swap-used">--</div>
                        <div class="stat-subvalue" id="swap-total">--</div>
                    </div>
                </div>
            </div>

            <div class="section">
                <div class="section-title">
                    <span class="process-icon">📋</span>
                    Processes (Top 50)
                </div>
                <table class="process-table">
                    <thead>
                        <tr>
                            <th>PID</th>
                            <th>Name</th>
                            <th>State</th>
                            <th>Memory</th>
                            <th>Threads</th>
                        </tr>
                    </thead>
                    <tbody id="process-list">
                        <tr><td colspan="5" class="loading">Loading...</td></tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div class="refresh-info">
            Auto-refreshing every 2 seconds
        </div>
    `;
}

// Update stats display
function updateStats(stats) {
    try {
        // Update header
        document.getElementById('header-cpu').textContent =
            `${stats.cpu.usage.toFixed(1)}%`;
        document.getElementById('header-memory').textContent =
            `${stats.memory.usedPercent.toFixed(1)}%`;
        document.getElementById('header-processes').textContent =
            stats.processes.length;
        document.getElementById('header-uptime').textContent =
            formatUptime(stats.uptime);

        // Update CPU stats
        document.getElementById('cpu-usage').textContent =
            `${stats.cpu.usage.toFixed(1)}%`;
        document.getElementById('cpu-progress').style.width =
            `${Math.min(stats.cpu.usage, 100)}%`;
        document.getElementById('cpu-cores').textContent =
            stats.cpu.cores;
        document.getElementById('cpu-model').textContent =
            stats.cpu.modelName || 'Unknown';

        // Update Memory stats
        document.getElementById('memory-used').textContent =
            formatBytes(stats.memory.used);
        document.getElementById('memory-used-percent').textContent =
            `${stats.memory.usedPercent.toFixed(1)}% of total`;
        document.getElementById('memory-progress').style.width =
            `${Math.min(stats.memory.usedPercent, 100)}%`;
        document.getElementById('memory-total').textContent =
            formatBytes(stats.memory.total);
        document.getElementById('memory-available').textContent =
            formatBytes(stats.memory.available);
        document.getElementById('memory-cached').textContent =
            formatBytes(stats.memory.cached);
        document.getElementById('swap-used').textContent =
            formatBytes(stats.memory.swapUsed);
        document.getElementById('swap-total').textContent =
            `of ${formatBytes(stats.memory.swapTotal)}`;

        // Update Process list
        const processList = document.getElementById('process-list');
        if (stats.processes && stats.processes.length > 0) {
            // Sort processes by memory usage
            const sortedProcesses = [...stats.processes].sort((a, b) => b.memory - a.memory);

            processList.innerHTML = sortedProcesses.map(proc => `
                <tr>
                    <td class="process-pid">${proc.pid}</td>
                    <td class="process-name">${proc.name}</td>
                    <td>
                        <span class="process-state ${getProcessStateClass(proc.state)}">
                            ${getProcessStateName(proc.state)}
                        </span>
                    </td>
                    <td>${formatBytes(proc.memory)}</td>
                    <td>${proc.threads}</td>
                </tr>
            `).join('');
        } else {
            processList.innerHTML = '<tr><td colspan="5" class="loading">No processes found</td></tr>';
        }

        lastUpdate = Date.now();
    } catch (error) {
        console.error('Error updating UI:', error);
    }
}

// Fetch and display system stats
async function fetchStats() {
    try {
        const stats = await GetSystemStats();
        updateStats(stats);
    } catch (error) {
        console.error('Error fetching stats:', error);
        document.querySelector('#app').innerHTML = `
            <div class="error">
                Failed to fetch system statistics: ${error.message || error}
            </div>
        `;
    }
}

// Initialize the app
function init() {
    createUI();
    fetchStats();

    // Update stats every 2 seconds
    if (updateInterval) {
        clearInterval(updateInterval);
    }
    updateInterval = setInterval(fetchStats, 2000);
}

// Start the app when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
