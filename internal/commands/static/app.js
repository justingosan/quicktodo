// Global state
let currentProject = '';
let tasks = [];
let ws = null;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;

// DOM Elements
const projectSelect = document.getElementById('project-select');
const kanbanBoard = document.getElementById('kanban-board');
const loading = document.getElementById('loading');
const error = document.getElementById('error');
const taskModal = document.getElementById('task-modal');
const taskForm = document.getElementById('task-form');
const addTaskBtn = document.getElementById('add-task-btn');
const deleteBtn = document.getElementById('delete-btn');
const cancelBtn = document.getElementById('cancel-btn');
const refreshBtn = document.getElementById('refresh-btn');
const modalClose = document.querySelector('.close');
const settingsBtn = document.getElementById('settings-btn');
const settingsModal = document.getElementById('settings-modal');
const settingsForm = document.getElementById('settings-form');
const settingsClose = document.getElementById('settings-close');
const settingsCancelBtn = document.getElementById('settings-cancel-btn');

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadProjects();
    setupEventListeners();
    connectWebSocket();
});

// WebSocket Connection
function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    try {
        ws = new WebSocket(wsUrl);
        
        ws.onopen = () => {
            console.log('WebSocket connected');
            reconnectAttempts = 0;
        };
        
        ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                handleWebSocketMessage(message);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };
        
        ws.onclose = () => {
            console.log('WebSocket disconnected');
            ws = null;
            // Attempt to reconnect
            if (reconnectAttempts < maxReconnectAttempts) {
                reconnectAttempts++;
                console.log(`Attempting to reconnect... (${reconnectAttempts}/${maxReconnectAttempts})`);
                setTimeout(connectWebSocket, Math.min(1000 * Math.pow(2, reconnectAttempts), 10000));
            }
        };
        
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    } catch (error) {
        console.error('Failed to create WebSocket connection:', error);
    }
}

function handleWebSocketMessage(message) {
    const { type, data, project } = message;
    
    // Only handle updates for the current project
    if (project && project !== currentProject) {
        return;
    }
    
    switch (type) {
        case 'task_created':
            handleTaskCreated(data);
            break;
        case 'task_updated':
            handleTaskUpdated(data);
            break;
        case 'task_deleted':
            handleTaskDeleted(data);
            break;
        default:
            console.log('Unknown WebSocket message type:', type);
    }
}

function handleTaskCreated(task) {
    // Add the new task to our local state
    tasks.push(task);
    
    // Re-render the tasks
    renderTasks();
    
    // Show a subtle notification
    showNotification(`New task created: ${task.title}`, 'info');
}

function handleTaskUpdated(task) {
    // Update the task in our local state
    const index = tasks.findIndex(t => t.id === task.id);
    if (index !== -1) {
        tasks[index] = task;
        
        // Re-render the tasks
        renderTasks();
        
        // Show a subtle notification
        showNotification(`Task updated: ${task.title}`, 'info');
    }
}

function handleTaskDeleted(taskData) {
    // Remove the task from our local state
    const index = tasks.findIndex(t => t.id === taskData.id);
    if (index !== -1) {
        tasks.splice(index, 1);
        
        // Re-render the tasks
        renderTasks();
        
        // Show a subtle notification
        showNotification(`Task deleted: ${taskData.title}`, 'info');
    }
}

function showNotification(message, type = 'info') {
    // Create a simple notification element
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: ${type === 'info' ? '#2196F3' : '#4CAF50'};
        color: white;
        padding: 12px 24px;
        border-radius: 4px;
        box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        z-index: 10000;
        font-size: 14px;
        opacity: 0;
        transform: translateX(100%);
        transition: all 0.3s ease;
    `;
    
    document.body.appendChild(notification);
    
    // Animate in
    setTimeout(() => {
        notification.style.opacity = '1';
        notification.style.transform = 'translateX(0)';
    }, 100);
    
    // Remove after 3 seconds
    setTimeout(() => {
        notification.style.opacity = '0';
        notification.style.transform = 'translateX(100%)';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, 3000);
}

// Event Listeners
function setupEventListeners() {
    projectSelect.addEventListener('change', handleProjectChange);
    addTaskBtn.addEventListener('click', () => openTaskModal());
    modalClose.addEventListener('click', closeModal);
    cancelBtn.addEventListener('click', closeModal);
    deleteBtn.addEventListener('click', handleDeleteTask);
    refreshBtn.addEventListener('click', refreshTasks);
    taskForm.addEventListener('submit', handleTaskSubmit);
    
    // Settings modal
    settingsBtn.addEventListener('click', openSettingsModal);
    settingsClose.addEventListener('click', closeSettingsModal);
    settingsCancelBtn.addEventListener('click', closeSettingsModal);
    settingsForm.addEventListener('submit', handleSettingsSubmit);
    
    // Close modals on background click
    taskModal.addEventListener('click', (e) => {
        if (e.target === taskModal) closeModal();
    });
    settingsModal.addEventListener('click', (e) => {
        if (e.target === settingsModal) closeSettingsModal();
    });
    
    // Load settings on startup
    loadSettings();
}

// API Functions
async function fetchAPI(url, options = {}) {
    try {
        const response = await fetch(url, options);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (err) {
        showError(`API Error: ${err.message}`);
        throw err;
    }
}

async function loadProjects() {
    try {
        const [projects, currentProjectInfo] = await Promise.all([
            fetchAPI('/api/projects'),
            fetchAPI('/api/current-project')
        ]);
        
        projectSelect.innerHTML = '<option value="">Select a project...</option>';
        projects.forEach(project => {
            const option = document.createElement('option');
            option.value = project.name;
            option.textContent = `${project.name} (${project.path})`;
            projectSelect.appendChild(option);
        });
        
        // Auto-select current project if detected
        if (currentProjectInfo.has_current_project && currentProjectInfo.current_project) {
            const currentProjectName = currentProjectInfo.current_project.name;
            projectSelect.value = currentProjectName;
            handleProjectChange();
            console.log(`Auto-selected current project: ${currentProjectName}`);
        }
        // Fall back to auto-select if only one project
        else if (projects.length === 1) {
            projectSelect.value = projects[0].name;
            handleProjectChange();
        }
    } catch (err) {
        console.error('Failed to load projects:', err);
    }
}

async function loadTasks() {
    if (!currentProject) return;
    
    showLoading(true);
    hideError();
    
    try {
        tasks = await fetchAPI(`/api/projects/${currentProject}/tasks`);
        renderTasks();
        kanbanBoard.style.display = 'flex';
    } catch (err) {
        console.error('Failed to load tasks:', err);
    } finally {
        showLoading(false);
    }
}

async function updateTask(taskId, updates) {
    try {
        const updated = await fetchAPI(`/api/projects/${currentProject}/tasks/${taskId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(updates)
        });
        
        const index = tasks.findIndex(t => t.id === taskId);
        if (index !== -1) {
            tasks[index] = updated;
        }
        return updated;
    } catch (err) {
        showError('Failed to update task');
        throw err;
    }
}

async function createTask(taskData) {
    try {
        const created = await fetchAPI(`/api/projects/${currentProject}/tasks`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(taskData)
        });
        tasks.push(created);
        return created;
    } catch (err) {
        showError('Failed to create task');
        throw err;
    }
}

async function deleteTask(taskId) {
    try {
        await fetchAPI(`/api/projects/${currentProject}/tasks/${taskId}`, {
            method: 'DELETE'
        });
        tasks = tasks.filter(t => t.id !== taskId);
    } catch (err) {
        showError('Failed to delete task');
        throw err;
    }
}

// UI Functions
function showLoading(show) {
    loading.style.display = show ? 'block' : 'none';
}

function showError(message) {
    error.textContent = message;
    error.style.display = 'block';
}

function hideError() {
    error.style.display = 'none';
}

function handleProjectChange() {
    currentProject = projectSelect.value;
    if (currentProject) {
        loadTasks();
    } else {
        kanbanBoard.style.display = 'none';
    }
}

function refreshTasks() {
    if (currentProject) {
        loadTasks();
    }
}

// Task Rendering
function renderTasks() {
    const columns = {
        pending: document.getElementById('pending-tasks'),
        in_progress: document.getElementById('in-progress-tasks'),
        done: document.getElementById('done-tasks')
    };
    
    // Clear all columns
    Object.values(columns).forEach(col => col.innerHTML = '');
    
    // Group tasks by status
    const grouped = {
        pending: [],
        in_progress: [],
        done: []
    };
    
    tasks.forEach(task => {
        if (grouped[task.status]) {
            grouped[task.status].push(task);
        }
    });
    
    // Render tasks and update counts
    Object.entries(grouped).forEach(([status, statusTasks]) => {
        const column = columns[status];
        const countElement = column.parentElement.querySelector('.task-count');
        countElement.textContent = statusTasks.length;
        
        statusTasks.forEach(task => {
            const card = createTaskCard(task);
            column.appendChild(card);
        });
    });
    
    // Setup drag and drop for columns
    Object.values(columns).forEach(column => {
        setupDropZone(column);
    });
}

function createTaskCard(task) {
    const card = document.createElement('div');
    card.className = 'task-card';
    card.draggable = true;
    card.dataset.taskId = task.id;
    
    // Format dates for display
    const createdDate = new Date(task.created_at).toLocaleDateString();
    const updatedDate = new Date(task.updated_at).toLocaleDateString();
    const createdTime = new Date(task.created_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
    const updatedTime = new Date(task.updated_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
    
    card.innerHTML = `
        <button class="copy-btn" title="Copy task details">📋</button>
        <div class="task-header">
            <h3>${escapeHtml(task.title)}</h3>
            <span class="task-id">#${task.id}</span>
        </div>
        ${task.description ? `<p class="task-description">${escapeHtml(task.description)}</p>` : ''}
        <div class="task-meta">
            <span class="priority ${task.priority}">${task.priority}</span>
            ${task.assigned_to ? `<span>@${escapeHtml(task.assigned_to)}</span>` : ''}
        </div>
        <div class="task-dates">
            <div class="date-info">
                <small>Created: ${createdDate} ${createdTime}</small>
            </div>
            ${task.created_at !== task.updated_at ? `<div class="date-info">
                <small>Updated: ${updatedDate} ${updatedTime}</small>
            </div>` : ''}
        </div>
    `;
    
    // Copy button functionality
    const copyBtn = card.querySelector('.copy-btn');
    copyBtn.addEventListener('click', (e) => {
        e.stopPropagation(); // Prevent opening modal
        copyTaskToClipboard(task);
    });
    
    // Click to open detail (but not on copy button)
    card.addEventListener('click', (e) => {
        if (!e.target.classList.contains('copy-btn')) {
            openTaskModal(task);
        }
    });
    
    // Drag events
    card.addEventListener('dragstart', handleDragStart);
    card.addEventListener('dragend', handleDragEnd);
    
    return card;
}

// Drag and Drop
let draggedElement = null;

function handleDragStart(e) {
    draggedElement = e.target;
    e.target.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/html', e.target.innerHTML);
}

function handleDragEnd(e) {
    e.target.classList.remove('dragging');
}

function setupDropZone(zone) {
    zone.addEventListener('dragover', handleDragOver);
    zone.addEventListener('drop', handleDrop);
    zone.addEventListener('dragenter', handleDragEnter);
    zone.addEventListener('dragleave', handleDragLeave);
}

function handleDragOver(e) {
    if (e.preventDefault) {
        e.preventDefault();
    }
    e.dataTransfer.dropEffect = 'move';
    return false;
}

function handleDragEnter(e) {
    e.target.closest('.tasks').classList.add('drag-over');
}

function handleDragLeave(e) {
    if (e.target.classList.contains('tasks')) {
        e.target.classList.remove('drag-over');
    }
}

async function handleDrop(e) {
    if (e.stopPropagation) {
        e.stopPropagation();
    }
    
    const dropZone = e.target.closest('.tasks');
    dropZone.classList.remove('drag-over');
    
    const newStatus = dropZone.parentElement.dataset.status;
    const taskId = draggedElement.dataset.taskId;
    
    try {
        await updateTask(taskId, { status: newStatus });
        renderTasks();
    } catch (err) {
        console.error('Failed to update task status:', err);
    }
    
    return false;
}

// Modal Functions
function openTaskModal(task = null) {
    const isNew = !task;
    
    document.getElementById('modal-title').textContent = isNew ? 'New Task' : 'Edit Task';
    document.getElementById('task-id').value = task?.id || '';
    document.getElementById('task-title').value = task?.title || '';
    document.getElementById('task-description').value = task?.description || '';
    document.getElementById('task-priority').value = task?.priority || 'medium';
    document.getElementById('task-status').value = task?.status || 'pending';
    document.getElementById('task-assigned').value = task?.assigned_to || '';
    
    deleteBtn.style.display = isNew ? 'none' : 'inline-block';
    taskModal.style.display = 'flex';
    
    // Focus on title
    setTimeout(() => document.getElementById('task-title').focus(), 100);
}

function closeModal() {
    taskModal.style.display = 'none';
    taskForm.reset();
}

async function handleTaskSubmit(e) {
    e.preventDefault();
    
    const taskId = document.getElementById('task-id').value;
    const taskData = {
        title: document.getElementById('task-title').value,
        description: document.getElementById('task-description').value,
        priority: document.getElementById('task-priority').value,
        status: document.getElementById('task-status').value,
        assigned_to: document.getElementById('task-assigned').value
    };
    
    try {
        if (taskId) {
            await updateTask(taskId, taskData);
        } else {
            await createTask(taskData);
        }
        renderTasks();
        closeModal();
    } catch (err) {
        console.error('Failed to save task:', err);
    }
}

async function handleDeleteTask() {
    const taskId = document.getElementById('task-id').value;
    if (!taskId) return;
    
    if (confirm('Are you sure you want to delete this task?')) {
        try {
            await deleteTask(taskId);
            renderTasks();
            closeModal();
        } catch (err) {
            console.error('Failed to delete task:', err);
        }
    }
}

// Copy Functionality
async function copyTaskToClipboard(task) {
    const settings = getSettings();
    const formatted = formatTaskForCopy(task, settings);
    
    try {
        await navigator.clipboard.writeText(formatted);
        showCopyFeedback(true);
    } catch (err) {
        // Fallback for older browsers
        try {
            const textArea = document.createElement('textarea');
            textArea.value = formatted;
            document.body.appendChild(textArea);
            textArea.select();
            document.execCommand('copy');
            document.body.removeChild(textArea);
            showCopyFeedback(true);
        } catch (fallbackErr) {
            showCopyFeedback(false);
            console.error('Failed to copy to clipboard:', fallbackErr);
        }
    }
}

function formatTaskForCopy(task, settings) {
    const prefix = settings.copyPrefix ? settings.copyPrefix + ' ' : '';
    
    switch (settings.copyFormat) {
        case 'detailed':
            return `${prefix}#${task.id} ${task.title}
Description: ${task.description || 'None'}
Status: ${task.status}
Priority: ${task.priority}
${task.assigned_to ? `Assigned to: ${task.assigned_to}` : ''}
Created: ${new Date(task.created_at).toLocaleDateString()}`;
            
        case 'markdown':
            return `${prefix}## #${task.id} ${task.title}
${task.description ? `\n${task.description}\n` : ''}
- **Status:** ${task.status}
- **Priority:** ${task.priority}
${task.assigned_to ? `- **Assigned:** ${task.assigned_to}` : ''}`;
            
        case 'simple':
        default:
            const desc = task.description ? ` - ${task.description}` : '';
            return `${prefix}#${task.id} ${task.title}${desc}`;
    }
}

function showCopyFeedback(success) {
    const feedback = document.createElement('div');
    feedback.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 12px 20px;
        background-color: ${success ? '#27ae60' : '#e74c3c'};
        color: white;
        border-radius: 4px;
        font-size: 14px;
        z-index: 10000;
        animation: slideIn 0.3s ease;
    `;
    feedback.textContent = success ? 'Task copied to clipboard!' : 'Failed to copy task';
    
    document.body.appendChild(feedback);
    
    setTimeout(() => {
        feedback.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => document.body.removeChild(feedback), 300);
    }, 2000);
}

// Settings Functions
function getSettings() {
    const defaults = {
        copyPrefix: 'Run `quicktodo context` and read CLAUDE.md before executing any commands.',
        copyFormat: 'simple'
    };
    
    try {
        const stored = localStorage.getItem('quicktodo-settings');
        return stored ? { ...defaults, ...JSON.parse(stored) } : defaults;
    } catch (err) {
        console.error('Failed to load settings:', err);
        return defaults;
    }
}

function saveSettings(settings) {
    try {
        localStorage.setItem('quicktodo-settings', JSON.stringify(settings));
        return true;
    } catch (err) {
        console.error('Failed to save settings:', err);
        return false;
    }
}

function loadSettings() {
    const settings = getSettings();
    document.getElementById('copy-prefix').value = settings.copyPrefix;
    document.getElementById('copy-format').value = settings.copyFormat;
}

function openSettingsModal() {
    loadSettings(); // Refresh settings when opening
    settingsModal.style.display = 'flex';
    setTimeout(() => document.getElementById('copy-prefix').focus(), 100);
}

function closeSettingsModal() {
    settingsModal.style.display = 'none';
    settingsForm.reset();
}

function handleSettingsSubmit(e) {
    e.preventDefault();
    
    const settings = {
        copyPrefix: document.getElementById('copy-prefix').value,
        copyFormat: document.getElementById('copy-format').value
    };
    
    if (saveSettings(settings)) {
        showCopyFeedback(true);
        closeSettingsModal();
    } else {
        showCopyFeedback(false);
    }
}

// Utility Functions
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Add CSS animations for copy feedback
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }
`;
document.head.appendChild(style);