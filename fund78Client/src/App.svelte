<script>
  let httpTopic = 'LOGON';
  let httpPayload = 'alice';
  let httpResponse = '';

  let wsTopic = 'TICK';
  let wsPayload = '12345';
  let wsStatus = 'Disconnected';
  let wsLog = '';
  let ws = null;

  async function sendHTTP() {
    try {
      const response = await fetch('http://localhost:8081/visitor', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ topic: httpTopic, payload: httpPayload })
      });

      const data = await response.json();
      httpResponse = `Status: ${response.status}\n${JSON.stringify(data, null, 2)}`;
    } catch (error) {
      httpResponse = `Error: ${error.message}`;
    }
  }

  function connectWS() {
    if (ws) {
      wsStatus = 'Already connected';
      return;
    }

    ws = new WebSocket('ws://localhost:8082/ws');

    ws.onopen = () => {
      wsStatus = 'Connected';
      addLog('Connected to WebSocket');
    };

    ws.onclose = () => {
      wsStatus = 'Disconnected';
      addLog('Disconnected from WebSocket');
      ws = null;
    };

    ws.onerror = (error) => {
      addLog(`Error: ${error}`);
    };

    ws.onmessage = (event) => {
      addLog(`Received: ${event.data}`);
    };
  }

  function disconnectWS() {
    if (ws) {
      ws.close();
    }
  }

  function sendWS() {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      addLog('Not connected. Click Connect first.');
      return;
    }

    const message = JSON.stringify({ topic: wsTopic, payload: wsPayload });
    ws.send(message);
    addLog(`Sent: ${message}`);
  }

  function addLog(message) {
    const timestamp = new Date().toLocaleTimeString();
    wsLog += `[${timestamp}] ${message}\n`;
  }
</script>

<h1>Fund78 Client</h1>

<div>
  <h2>HTTP POST /visitor</h2>
  <input type="text" bind:value={httpTopic} placeholder="Topic (e.g., LOGON)">
  <input type="text" bind:value={httpPayload} placeholder="Payload (e.g., alice)">
  <button on:click={sendHTTP}>Send HTTP</button>
  <pre>{httpResponse}</pre>
</div>

<hr>

<div>
  <h2>WebSocket /ws</h2>
  <button on:click={connectWS}>Connect</button>
  <button on:click={disconnectWS}>Disconnect</button>
  <div>Status: {wsStatus}</div>
  <br>
  <input type="text" bind:value={wsTopic} placeholder="Topic (e.g., TICK)">
  <input type="text" bind:value={wsPayload} placeholder="Payload (e.g., 12345)">
  <button on:click={sendWS}>Send WebSocket</button>
  <pre>{wsLog}</pre>
</div>
