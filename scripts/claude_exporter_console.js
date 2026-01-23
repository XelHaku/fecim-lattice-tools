/*
 * Claude.ai Chat Exporter (Console Version)
 * 
 * INSTRUCTIONS:
 * 1. Open the Developer Tools on claude.ai (F12 or Ctrl+Shift+I).
 * 2. Go to the "Console" tab.
 * 3. If you see a warning, type 'allow pasting' and hit Enter.
 * 4. Paste ALL of this code into the console and hit Enter.
 * 5. Buttons "Export Chat" or "Export All Chats" will appear in the bottom right.
 */

(function () {
    'use strict';
 
    const API_BASE_URL = 'https://claude.ai/api';
 
    // Function to make API requests (using standard fetch for Console compatibility)
    async function apiRequest(method, endpoint, data = null, headers = {}) {
        const url = `${API_BASE_URL}${endpoint}`;
        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                ...headers,
            },
            body: data ? JSON.stringify(data) : null
        };

        if (method === 'GET' || method === 'HEAD') {
            delete options.body;
        }

        const response = await fetch(url, options);
        if (!response.ok) {
            throw new Error(`API request failed with status ${response.status}`);
        }
        return await response.json();
    }
 
    // Function to get the organization ID
    async function getOrganizationId() {
        const organizations = await apiRequest('GET', '/organizations');
        return organizations[0].uuid;
    }
 
    // Function to get all conversations
    async function getAllConversations(orgId) {
        return await apiRequest('GET', `/organizations/${orgId}/chat_conversations`);
    }
 
    // Function to get conversation history
    async function getConversationHistory(orgId, chatId) {
        return await apiRequest('GET', `/organizations/${orgId}/chat_conversations/${chatId}`);
    }
 
    // Function to download data as a file
    function downloadData(data, filename, format) {
        return new Promise((resolve, reject) => {
            let content = '';
            if (format === 'json') {
                content = JSON.stringify(data, null, 2);
            } else if (format === 'txt') {
                content = convertToTxtFormat(data);
            }
            const blob = new Blob([content], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            a.style.display = 'none';
            document.body.appendChild(a);
            a.click();
            setTimeout(() => {
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
                resolve();
            }, 100);
        });
    }
 
    // Function to convert conversation data to TXT format
    function convertToTxtFormat(data) {
        let txtContent = '';
        data.chat_messages.forEach((message) => {
            const sender = message.sender === 'human' ? 'User' : 'Claude';
            txtContent += `${sender}:\n${message.text}\n\n`;
        });
        return txtContent.trim();
    }
 
    // Function to export a single chat
    async function exportChat(orgId, chatId, format, showAlert = true) {
        try {
            const chatData = await getConversationHistory(orgId, chatId);
            const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
            const filename = `${chatData.name}_${timestamp}.${format}`;
            await downloadData(chatData, filename, format);
            if (showAlert) {
                alert(`Chat exported successfully in ${format.toUpperCase()} format!`);
            }
        } catch (error) {
            console.error(error);
            alert('Error exporting chat. Check console for details.');
        }
    }
 
    // Function to export all chats
    async function exportAllChats(format) {
        try {
            const orgId = await getOrganizationId();
            const conversations = await getAllConversations(orgId);
            let count = 0;
            for (const conversation of conversations) {
                console.log(`Exporting ${conversation.name}...`);
                await exportChat(orgId, conversation.uuid, format, false);
                count++;
            }
            alert(`Successfully exported ${count} chats in ${format.toUpperCase()} format!`);
        } catch (error) {
            console.error(error);
            alert('Error exporting all chats. Check console for details.');
        }
    }
 
    // Function to create a button
    function createButton(text, onClick) {
        const button = document.createElement('button');
        button.textContent = text;
        button.style.cssText = `
            position: fixed;
            bottom: 20px;
            right: 20px;
            padding: 10px 20px;
            background-color: #d93d3d;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            z-index: 9999;
            box-shadow: 0 2px 5px rgba(0,0,0,0.2);
        `;
        button.addEventListener('click', onClick);
        document.body.appendChild(button);
    }
 
    // Function to remove existing export buttons
    function removeExportButtons() {
        const existingButtons = document.querySelectorAll('button[style*="position: fixed"]'); // Basic selector
        // More specific check to essentially only remove buttons we likely added or similar position
        existingButtons.forEach((button) => {
            if (button.textContent.includes('Export')) {
                 button.remove();
            }
        });
    }
 
    // Function to initialize the export functionality
    async function initExportFunctionality() {
        removeExportButtons();
        const currentUrl = window.location.href;
        
        try {
            const orgId = await getOrganizationId();
            
            if (currentUrl.includes('/chat/')) {
                const urlParts = currentUrl.split('/');
                const chatId = urlParts[urlParts.length - 1];
                createButton('Export This Chat', async () => {
                    const format = prompt('Enter the export format (json or txt):', 'json');
                    if (format === 'json' || format === 'txt') {
                        await exportChat(orgId, chatId, format);
                    }
                });
            } else {
                // Assuming list view or generic
                createButton('Export All Chats', async () => {
                    const format = prompt('Enter the export format (json or txt):', 'json');
                    if (format === 'json' || format === 'txt') {
                        await exportAllChats(format);
                    }
                });
            }
        } catch (e) {
            console.error("Failed to init export functionality. Are you logged in?", e);
        }
    }
 
    // Run immediately
    initExportFunctionality();
    
    console.log("Claude Exporter loaded! Look for the button in the bottom right.");
})();
