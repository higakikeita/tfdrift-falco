#!/usr/bin/env node

// Simple WebSocket test client for TFDrift-Falco
const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', function open() {
  console.log('✅ WebSocket connected');

  // Subscribe to all topics
  console.log('📤 Subscribing to "all" topic...');
  ws.send(JSON.stringify({
    type: 'subscribe',
    topic: 'all'
  }));

  // Subscribe to drifts specifically
  setTimeout(() => {
    console.log('📤 Subscribing to "drifts" topic...');
    ws.send(JSON.stringify({
      type: 'subscribe',
      topic: 'drifts'
    }));
  }, 1000);

  // Send ping
  setTimeout(() => {
    console.log('📤 Sending ping...');
    ws.send(JSON.stringify({
      type: 'ping'
    }));
  }, 2000);

  // Keep alive for 30 seconds to receive events
  setTimeout(() => {
    console.log('⏱️  Test completed, closing connection...');
    ws.close();
  }, 30000);
});

ws.on('message', function message(data) {
  try {
    const msg = JSON.parse(data);
    console.log('📥 Received:', JSON.stringify(msg, null, 2));
  } catch (e) {
    console.log('📥 Received (raw):', data.toString());
  }
});

ws.on('error', function error(err) {
  console.error('❌ WebSocket error:', err.message);
});

ws.on('close', function close() {
  console.log('🔌 WebSocket connection closed');
  process.exit(0);
});
