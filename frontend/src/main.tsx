import ReactDOM from 'react-dom/client';
import { StrictMode } from 'react';
import { RouterProvider } from '@tanstack/react-router';
import './styles/globals.css';
import { router } from './router';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
);
