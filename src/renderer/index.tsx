import React from 'react';
import { createRoot } from 'react-dom/client';

const App = () => <h1>AUDiotMonitor: Hello World</h1>;

const container = document.getElementById('root');
if (container) {
  createRoot(container).render(<App />);
}
