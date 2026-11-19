import { createRoot } from 'react-dom/client';
import { App } from './App';
import './styles.css';
import './workspace-overrides.css';
import './status-menu.css';

createRoot(document.getElementById('root')!).render(<App />);
