let config;

async function getConfig() {
  const res = await fetch('/config.json');
  return await res.json();
}

async function log(...msg) {
  if (config?.client?.debug) {
    console.log(new Date().toISOString(), 'webrtc', ...msg);
  }
}

async function connectToStream(streamId, videoElement) {
  if (!config) config = await getConfig();
  const streamName = streamId || config.client.defaultStream;
  const videoElem = typeof videoElement === 'string' ? document.getElementById(videoElement) : videoElement;
    
  const mediaStream = new MediaStream();
  console.log('yrdy' + config.server.iceServers);
  const peerConnection = new RTCPeerConnection({
    iceServers: config.server.iceServers?.map(server => ({ urls: server })) || [{ urls: 'stun:stun.l.google.com:19302' }]
  });
  
  // Setup data channel
  const dataChannel = peerConnection.createDataChannel(streamName);
dataChannel.onopen = () => {
  log('Data channel open');
  const pingInterval = setInterval(() => {
    if (dataChannel.readyState === 'open') {
      dataChannel.send('ping');
    } else {
      clearInterval(pingInterval);
      log('Data channel no longer open, stopping pings');
    }
  }, 1000);
};
  
  // Handle incoming tracks
  peerConnection.ontrack = (event) => {
    log(`Received ${event.track.kind} track`);
    mediaStream.addTrack(event.track);
    if (videoElem instanceof HTMLVideoElement) {
      videoElem.srcObject = mediaStream;
    }
  };
  
  // Default transceivers - we'll get proper info from the server response
  peerConnection.addTransceiver('audio', { direction: 'recvonly' });
  peerConnection.addTransceiver('video', { direction: 'recvonly' });
  
  // Create and set local description
  const offer = await peerConnection.createOffer();
  await peerConnection.setLocalDescription(offer);
  
  try {
    // Send offer to server
    //  log(`LOCAL SDP to stream: ${peerConnection.localDescription.sdp}`);

    const response = await fetch(`http://${location.hostname}${config.server.encoderPort}/stream/webrtc/${streamName}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        data: btoa(peerConnection.localDescription.sdp)
      })
    });
    
    if (!response.ok) {
      throw new Error(`Server error: ${response.status}`);
    }
    
    const { answer, codecs } = await response.json();
    log(`Stream has ${codecs.length} codecs:`, codecs);
    
    // Important fix: decode base64 answer
    const decodedAnswer = atob(answer);
    log('Decoded SDP answer');
    
    // Set remote description from decoded answer
    await peerConnection.setRemoteDescription({
      type: 'answer',
      sdp: decodedAnswer
    });
    
    log('Connection established successfully');
    return peerConnection;
  } catch (error) {
    log('Connection failed:', error);
    peerConnection.close();
    throw error;
  }
}

window.onload = () => connectToStream('reowhite', 'videoElem');