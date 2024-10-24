import React from 'react';
import './NavBar.css';
import { Link } from 'react-router-dom';
import '@fortawesome/fontawesome-free/css/all.min.css';

const Navbar = () => {
  return (
    <nav className="navbar">
        <div className="logo">
            <h2>Sistema De Archivos EX2 y EX3</h2>
        </div>
        <ul className="nav-links">
            <li>
                <Link to="/execution">
                    <i className="fas fa-play"></i> Ejecución
                </Link>
            </li>
            <li>
                <Link to="/login">
                    <i className="fas fa-sign-in-alt"></i> Iniciar sesión
                </Link>
            </li>
            <li>
                <Link to="/visualizador">
                    <i className="fas fa-eye"></i> Visualizador
                </Link>
            </li>
        </ul>
    </nav>
  );
};

export default Navbar;
