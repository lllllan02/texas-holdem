import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import Table from './pages/Table'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/table" element={<Table />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App