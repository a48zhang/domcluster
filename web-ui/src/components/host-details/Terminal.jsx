import React, { useEffect, useRef } from 'react';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { Card } from '../../common';

const Terminal = ({ nodeId, nodeInfo, visible, onToggle }) => {
  const terminalRef = useRef(null);
  const xtermRef = useRef(null);
  const fitAddonRef = useRef(null);
  const wsRef = useRef(null);

  useEffect(() => {
    if (!visible || !terminalRef.current || xtermRef.current) return;

    const term = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Consolas, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#f8f8f8',
      },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);
    fitAddon.fit();

    xtermRef.current = term;
    fitAddonRef.current = fitAddon;

    // Connect to WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/terminal/ws?node_id=${nodeId}`;
    
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      term.writeln('Connected to ' + (nodeInfo?.name || nodeId));
      term.writeln('Type commands and press Enter to execute');
      term.writeln('');
      term.write('$ ');
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'output' && msg.data) {
          term.write(msg.data);
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    ws.onerror = (error) => {
      term.writeln('\r\nWebSocket error occurred');
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      term.writeln('\r\nConnection closed');
    };

    let currentLine = '';

    term.onData((data) => {
      const code = data.charCodeAt(0);

      if (code === 13) { // Enter key
        term.write('\r\n');
        if (currentLine.trim() && ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({
            type: 'input',
            data: currentLine.trim()
          }));
        }
        currentLine = '';
        term.write('$ ');
      } else if (code === 127) { // Backspace
        if (currentLine.length > 0) {
          currentLine = currentLine.slice(0, -1);
          term.write('\b \b');
        }
      } else if (code >= 32) { // Printable characters
        currentLine += data;
        term.write(data);
      }
    });

    const handleResize = () => {
      if (fitAddonRef.current && visible) {
        fitAddonRef.current.fit();
      }
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      if (wsRef.current) {
        wsRef.current.close();
      }
      term.dispose();
      xtermRef.current = null;
    };
  }, [visible, nodeId, nodeInfo]);

  const headerRight = (
    <button className="btn btn-toggle" onClick={onToggle}>
      {visible ? '隐藏' : '显示'} Terminal
    </button>
  );

  return (
    <Card title="Web Terminal" headerRight={headerRight}>
      {visible && (
        <div className="terminal-container">
          <div ref={terminalRef} className="terminal"></div>
        </div>
      )}
    </Card>
  );
};

export default Terminal;
