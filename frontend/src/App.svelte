<script>
    import { onMount, tick } from 'svelte';

    const SESSION_KEY = 'ri_active_session_key';
    const ENDPOINTS = {
        start: '/api/create',
        extend: '/api/extend/',
        data: '/api/session/',
        end: '/api/delete/',
        clear: '/api/clear/',
        log_request: '/api/request/'
    }
    const BUFFER_SIZE = 5;
    const REFRESH_INTERVAL_SECONDS = 5;

    let base_url = import.meta.env.VITE_BASE_URL;
    let session_key = null;
    let session_data = [];
    let polling_service = null;
    let expandedRequests = [];
    let requestHeights = {};

    const currentProtocol = window.location.protocol;
    const currentHost = window.location.host;
    const currentPath = currentProtocol + '//' + currentHost;

    function getSessionFromURL() {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get('session');
    }

    function resetAppState() {
        session_key = null;
        session_data = [];
        expandedRequests = [];
        requestHeights = {};
        if (polling_service) {
            clearInterval(polling_service);
            polling_service = null;
        }
        localStorage.removeItem(SESSION_KEY);
        window.history.replaceState({}, document.title, '/');
    }

    onMount(async () => {
        if (!base_url || base_url.length === 0 || base_url === '/') {
            base_url = currentPath;
        }

        let _session_key = getSessionFromURL() || localStorage.getItem(SESSION_KEY);
        await checkValidSession(_session_key);

        if (session_data.length > 0) {
            expandedRequests = [0];
            await tick();
            updateRequestHeight(0);
        }
    });

    async function serverRequest(url, method, body = null) {
        let _request = {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            }
        }

        if ( body ) {
            _request.body = JSON.stringify(body);
        }

        let resp = await fetch(url, _request);
        let response = null;
        try {
            response = await resp.json();
        } catch (error) {
            console.error('request parse error', error);
        }

        if ( resp.status !== 200 ) {
            console.error('Error in request', resp);
            if ( response.error) {
                alert(response.error);
                response = null;
            } else {
                alert('Error in request');
                response = null;
            }
        }

        return response;
    }

    async function checkValidSession(_session_key) {
        if (!_session_key) {
            resetAppState();
            return;
        }

        let data = await serverRequest(`${base_url}${ENDPOINTS.data}${_session_key}`, 'GET');
        if (data) {
            session_key = _session_key;
            session_data = data;
            polling_service = setInterval(async () => {
                await getSessionData();
            }, REFRESH_INTERVAL_SECONDS * 1000);
        } else {
            resetAppState();
        }
    }

    async function startSession() {
        if ( session_key ) {
            alert('Delete the current session to start a new one');
            return;
        }

        let data = await serverRequest(`${base_url}${ENDPOINTS.start}`, 'GET');
        if ( data && data.session_id && data.session_id.length > 0) {
            session_key = data.session_id;
            localStorage.setItem(SESSION_KEY, session_key);
        } else {
            console.error('Error starting session', data);
        }

    }
    
    async function extendSession() {

        let data = await serverRequest(`${base_url}${ENDPOINTS.extend}${session_key}`, 'GET');
        if ( data ){
            console.log('Session extended', data);
            alert('Session extended');
        }
    }

    
    async function endSession() {
        let data = await serverRequest(`${base_url}${ENDPOINTS.end}${session_key}`, 'GET');
        if (data) {
            resetAppState();
        }
    }
    
    async function getSessionData() {

        let data = await serverRequest(`${base_url}${ENDPOINTS.data}${session_key}`, 'GET');
        if ( data ) {
            session_data = data;
        } 
    }
    
    async function clearSessionData() {
        let data = await serverRequest(`${base_url}${ENDPOINTS.clear}${session_key}`, 'GET');
        if ( data ) {
            session_data = [];
        } 
    }

    async function simulateRequest() {
        let body = {
            "message": 'This is a test request'
        }

        let data = await serverRequest(`${base_url}${ENDPOINTS.log_request}${session_key}`, 'POST', body);
        if ( data ) {
            console.log('Request logged', data);
            // update local session data
            await getSessionData();
        }
    }

    function formatRequest(request) {
        const lines = request.split('\n');
        const formattedLines = lines.map((line, index) => {
            if (index === 0) {
                return `<span class="method-url">${line}</span>`;
            } else if (line.includes(':')) {
                const [key, ...value] = line.split(':');
                return `<span class="header-key">${key}:</span><span class="header-value">${value.join(':')}</span>`;
            } else {
                return line;
            }
        });
        return formattedLines.join('\n');
    }

    function copyToClipboard(text) {
        navigator.clipboard.writeText(text).then(() => {
            alert('Copied to clipboard!');
        }, (err) => {
            console.error('Could not copy text: ', err);
        });
    }

    async function toggleRequest(index) {
        if (expandedRequests.includes(index)) {
            expandedRequests = expandedRequests.filter(i => i !== index);
        } else {
            expandedRequests = [...expandedRequests, index];
            await tick();
            updateRequestHeight(index);
        }
    }

    function updateRequestHeight(index) {
        const content = document.getElementById(`request-content-${index}`);
        if (content) {
            requestHeights[index] = content.scrollHeight + 'px';
        }
    }

    $: console.log('data >', session_data);
</script>

<main>
    <div class="container">
        <header>
            <h1>Request Inspector</h1>
            <p class="subtitle">Endpoint: {currentPath}</p>
        </header>

        {#if !session_key}
            <div class="welcome-section">
                <h2>Welcome to Request Inspector</h2>
                <p>To start a new session, click the 'Start Session' button below.</p>
                <button class="btn btn-primary" on:click={startSession}>Start Session</button>
            </div>        
        {:else}
            <div class="session-info">
                <h2>Active Session</h2>
                <p>Your Session ID: <strong>{session_key}</strong></p>
                <p>Send requests to: <code>{base_url}{ENDPOINTS.log_request}{session_key}</code></p>
            </div>
            <div class="actions">
                <h3>Actions</h3>
                <div class="button-group">
                    <button class="btn" on:click={simulateRequest}>Simulate Request</button>
                    <button class="btn" on:click={getSessionData}>Refresh Data</button>
                    <button class="btn" on:click={extendSession}>Extend Session</button>
                    <button class="btn" on:click={clearSessionData}>Clear Data</button>
                    <button class="btn btn-danger" on:click={endSession}>End Session</button>
                </div>
            </div>
            <div class="requests-wrapper">
                <h3>Logged Requests <span class="badge">{session_data.length} / {BUFFER_SIZE}</span></h3>
                <div class="request-container">
                    {#each session_data as request, index}
                        <div class="request">
                            <div class="request-header" on:click={() => toggleRequest(index)}>
                                Request {index + 1}
                                <span class="expand-icon">{expandedRequests.includes(index) ? '▼' : '▶'}</span>
                            </div>
                            <div class="request-content" 
                                 style="max-height: {expandedRequests.includes(index) ? requestHeights[index] : '0px'};">
                                <div id="request-content-{index}">
                                    <pre>
                                        {@html formatRequest(request)}
                                    </pre>
                                    <button class="btn btn-small copy-btn" on:click|stopPropagation={() => copyToClipboard(request)}>
                                        Copy
                                    </button>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            </div>
        {/if}
    </div>
</main>

<style>
    :global(body) {
        font-family: Arial, sans-serif;
        line-height: 1.6;
        color: #333;
        background-color: #f4f4f4;
        margin: 0;
        padding: 0;
    }

    .container {
        max-width: 1200px;
        margin: 0 auto;
        padding: 20px;
    }

    header {
        background-color: #007bff;
        color: white;
        padding: 20px;
        border-radius: 5px;
        margin-bottom: 20px;
    }

    h1, h2, h3 {
        margin-bottom: 10px;
    }

    .subtitle {
        font-size: 0.9em;
        opacity: 0.8;
    }

    .welcome-section, .session-info, .actions, .requests-wrapper {
        background-color: white;
        border-radius: 5px;
        padding: 20px;
        margin-bottom: 20px;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }

    .btn {
        padding: 10px 15px;
        border: none;
        border-radius: 5px;
        cursor: pointer;
        font-size: 14px;
        transition: background-color 0.3s;
    }

    .btn-primary {
        background-color: #007bff;
        color: white;
    }

    .btn-primary:hover {
        background-color: #0056b3;
    }

    .btn-danger {
        background-color: #dc3545;
        color: white;
    }

    .btn-danger:hover {
        background-color: #c82333;
    }

    .button-group {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
    }

    .badge {
        background-color: #17a2b8;
        color: white;
        padding: 3px 7px;
        border-radius: 10px;
        font-size: 0.8em;
    }

    .request-container {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .request {
        background-color: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 5px;
        overflow: hidden;
    }

    .request-header {
        padding: 10px 15px;
        background-color: #e9ecef;
        cursor: pointer;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .request-content {
        overflow: hidden;
        transition: max-height 0.3s ease-out;
    }

    .request-content > div {
        padding: 15px;
        position: relative;
    }

    .expand-icon {
        font-size: 12px;
    }

    .request pre {
        margin: 0;
        white-space: pre-wrap;
        word-wrap: break-word;
    }

    .method-url {
        color: #e83e8c;
        font-weight: bold;
    }

    .header-key {
        color: #0056b3;
    }

    .header-value {
        color: #28a745;
    }

    .btn-small {
        position: absolute;
        top: 10px;
        right: 10px;
        padding: 5px 10px;
        font-size: 12px;
        background-color: #6c757d;
        color: white;
    }

    .btn-small:hover {
        background-color: #5a6268;
    }

    .copy-btn {
        position: absolute;
        top: 10px;
        right: 10px;
        z-index: 1;
    }

    code {
        background-color: #f8f9fa;
        padding: 2px 4px;
        border-radius: 4px;
        font-family: monospace;
    }

    @media (max-width: 768px) {
        .request {
            width: 100%;
        }
    }
</style>