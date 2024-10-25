import React, { useState } from 'react';
import './Login.css';
import '@fortawesome/fontawesome-free/css/all.min.css';
import Swal from "sweetalert2";

const Login = ({ EnLogin }) => {
  const [IdUsuario, setIdUsuario] = useState('');
  const [Contrasena, setContrasena] = useState('');
  const [IdParticion, setIdParticion] = useState('');

  const handleLogin = async () => {
    const comando = `login -user="${IdUsuario}" -pass="${Contrasena}" -id="${IdParticion}"`;
    try {
        const response = await fetch('http://localhost:8080/AnalizadorGo/ProcesarComando', {
            method: 'POST',
            headers: {
                'Content-Type': 'text/plain',
            },
            body: comando.toLocaleLowerCase(),
        });
        const respuesta = await response.text();
        
        if (respuesta.includes("Usuario logueado con éxito en la partición")) {
            Swal.fire(respuesta, "Bienvenido A La aplicación", "success");
            localStorage.setItem("UsuarioLogueado", IdUsuario);

            if (EnLogin) {
              EnLogin(IdUsuario);
            }

        } else {
            Swal.fire(respuesta, "Error Al Iniciar Sesión", "error");
        }
        
    } catch (error) {
        Swal.fire("Error Al Enviar Texto: " + error, "Error Desconocido", "error");
    }
}


  return (
    <div className="login-container">
      <h2 className="login-title">Login</h2>
      <div className="login-mb">
        <label className="login-label">
          <i className="fas fa-id-badge"></i> ID Partición
        </label>
        <input
          type="text"
          className="login-input"
          placeholder="Ingrese el ID de la partición"
          value={IdParticion}
          onChange={(e) => setIdParticion(e.target.value)}
        />
      </div>
      <div className="login-mb">
        <label className="login-label">
          <i className="fas fa-user"></i> Usuario
        </label>
        <input
          type="text"
          className="login-input"
          placeholder="Ingrese su usuario"
          value={IdUsuario}
          onChange={(e) => setIdUsuario(e.target.value)}
        />
      </div>
      <div className="login-mb">
        <label className="login-label">
          <i className="fas fa-lock"></i> Contraseña
        </label>
        <input
          type="password"
          className="login-input"
          placeholder="Ingrese su contraseña"
          value={Contrasena}
          onChange={(e) => setContrasena(e.target.value)}
        />
      </div>
      <button className="login-btn" onClick={handleLogin}>
        <i className="fas fa-sign-in-alt"></i> Iniciar sesión
      </button>
    </div>
  );
};

export default Login;
