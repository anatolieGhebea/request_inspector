<script>
    import { onMount } from 'svelte';

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

    const currentProtocol = window.location.protocol;
    const currentHost = window.location.host;
    const currentPath = currentProtocol + '//' + currentHost;

    onMount(async  () => {
        if ( !base_url || base_url.length === 0 || base_url === '/') {
            base_url = currentPath;
        }

        let _session_key = localStorage.getItem(SESSION_KEY);
        // check if session is still active
       await checkValidSession(_session_key);
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
        if ( !_session_key ) {
            return;
        }

        let data = await serverRequest(`${base_url}${ENDPOINTS.data}${_session_key}`, 'GET');
        if ( data ) {
            session_key = _session_key;
            session_data = data;
            polling_service = setInterval(async () => {
                await getSessionData();
            }, REFRESH_INTERVAL_SECONDS * 1000);
        } else {
            session_key = null;
            localStorage.removeItem(SESSION_KEY);
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
        if ( data ) {
            clearInterval(polling_service);
            session_key = null;
            localStorage.removeItem(SESSION_KEY);
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

    $: console.log('data >', session_data);
</script>

<main>
    {currentPath}
{#if !session_key}
<div>
    <h1>Welcome to Reqeust Inspector</h1>
    <h3>To start a new session click the 'Start Session'</h3>
    <button class="start_session" on:click={startSession}>Start Session</button>
</div>        
{:else}
<div>
    <div>
        <h1>Your Session_ID is: <b>{session_key}</b> </h1>
        You can now send your requesto to the following endpoint: <strong>{base_url}{ENDPOINTS.log_request}{session_key}</strong>
    </div>
    <div class="actions">
        <h2>
            Actions
        </h2>
        <div>
            <button class="simulate_request" on:click={simulateRequest}>Simulate Request</button>
            <button class="get_session_data" on:click={getSessionData}>Get Session Data</button>
            <button class="extend_session" on:click={extendSession}>Extend Session</button>
            <button class="clear_session_data" on:click={clearSessionData}>Clear Session Data</button>
            <button class="end_session" on:click={endSession}>End Session</button>
        </div>
    </div>
    <div class="requests_wrapper">
        <h2>Current logged request <small>{session_data.length} / {BUFFER_SIZE}</small> </h2>
        <div class="request_container">
            {#each session_data as request}
                <div class="request" style="padding: 1rem;">
                    {request}
                </div>
            {/each}
        </div>
    </div>


</div>
{/if}
</main>

<style>
    .requests_wrapper {
        margin-top: 1rem;
        width: 100%;
        max-width: 100%;
        overflow: auto;
    }

    .request_container {
        /* width: 100%; */
        display: flex;
        flex-wrap: nowrap;
        justify-content: flex-start;
        align-items: flex-start;
    }
    .request {
        width: 400px;
        overflow: auto;
        margin: .25rem;
        padding: .5rem;
        border: 1px dashed #ccc;
        border-radius: .35rem;
    }
</style>