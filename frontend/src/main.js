import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Navbar from './components/NavBar';
import Ejecutador from './pages/Ejecutador';
import Login from './pages/Login';
import Visualizador from './pages/Visualizador';

const main = () => {
    return (
        <Router>
        <Navbar />
        <Routes>
            <Route path="/" element={<Ejecutador />} />
            <Route path="/execution" element={<Ejecutador />} />
            <Route path="/login" element={<Login />} />
            <Route path="/visualizador" element={<Visualizador />} />
        </Routes>
        </Router>
    );
}

export default main;