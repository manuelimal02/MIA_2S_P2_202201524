import React from 'react';
import './NavBar.css';
import { Link } from 'react-router-dom';
import '@fortawesome/fontawesome-free/css/all.min.css';

const Navbar = () => {
  return (
    <nav className="navbar-principal">
        <div className="logo">
            <h2>Sistema De Archivos EX2-EX3</h2>
        </div>
        <ul className="navbar-enlaces">
            <li><Link to="/execution"><i className="fas fa-play"></i> Ejecución</Link></li>
            <li><Link to="/login"><i className="fas fa-address-card"></i> Iniciar sesión</Link></li>
            <li><Link to="/visualizador"><i className="fas fa-tv"></i> Visualizador</Link></li>
        </ul>
    </nav>
  );
};

export default Navbar;
