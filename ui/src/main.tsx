import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import AppFinal from './App-final.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AppFinal />
  </StrictMode>,
)
