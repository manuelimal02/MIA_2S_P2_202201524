import './Ejecutador.css';
import '@fortawesome/fontawesome-free/css/all.min.css';
import React, { useEffect, useState } from 'react';
import Swal from "sweetalert2";

function App() {
  const [TextoEntrada, SetTextoEntrada] = useState('');
  const [TextoSalida, SetTextoSalida] = useState('');
  const [UsuarioL, setUsuario] = useState('');

  // Recuperar el UsuarioL desde localStorage
  useEffect(() => {
    const UsuarioLogueado = localStorage.getItem('UsuarioLogueado');
    setUsuario(UsuarioLogueado || 'No Existe Usuario Logueado');
  }, []);

  // Función para abrir el explorador de archivos
  const SeleccionadorArchivos = () => {
    const ArchivoEntrante = document.createElement('input');
    ArchivoEntrante.type = 'file';
    ArchivoEntrante.accept = '.smia';
    ArchivoEntrante.onchange = HandleSeleccionarArchivos;
    ArchivoEntrante.click();
  };

  const HandleSeleccionarArchivos = (event) => {
    const ArchivoEntrante = event.target.files[0];
    if (ArchivoEntrante && ArchivoEntrante.name.endsWith('.smia')) {
      const reader = new FileReader();
      reader.onload = (e) => {
        SetTextoEntrada(e.target.result);
      };
      reader.readAsText(ArchivoEntrante);
    } else {
      Swal.fire("Se debe seleccionar un archivo con la extensión '.smia'", "Error", "error");
    }
  };

  // Función para enviar el texto al backend "todos los comandos"
  const HandleProcesarComando = async () => {
    try {
      const response = await fetch('http://localhost:8080/AnalizadorGo/ProcesarComando', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: TextoEntrada,
      });
      const respuesta = await response.text();
      SetTextoSalida(respuesta);  
    } catch (error) {
      Swal.fire("Error Al Enviar Texto: " + error, "Error Desconocido", "error");
      SetTextoSalida('Error al procesar los comandos de entrada.'); 
    }
  };

  // Función para enviar el texto al backend "logout"
  const HandleCerrarSesion = async () => {
    try {
      const response = await fetch('http://localhost:8080/AnalizadorGo/ProcesarComando', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: "logout",
      });
      const respuesta = await response.text();

      if (respuesta.includes("Sesión cerrada con éxito de la partición")) {
        Swal.fire(respuesta, "Cerrando Sesión", "success");
        localStorage.removeItem("UsuarioLogueado");
        setTimeout(() => {
          window.location.href = '/';
        }, 2000);
      } else {
        Swal.fire(respuesta, "Error Al Cerrar Sesión", "error");
      }
      SetTextoSalida(respuesta);

    } catch (error) {
      Swal.fire("Error Al Enviar Texto: "+error, "Error Desconocido", "error");
      SetTextoSalida('Error al procesar los comandos de entrada.'); 
    }
  };

  return (
    <div className="container">
      <div className="buttons">
        <button id="selectBtn" onClick={SeleccionadorArchivos}>
          <i className="fas fa-folder-open"></i> Seleccionar
        </button>
        <button id="executeBtn" onClick={HandleProcesarComando}>
          <i className="fas fa-play"></i> Ejecutar
        </button>

        <button id="logoutBtn" onClick={HandleCerrarSesion}>
          <i className="fas fa-right-from-bracket"></i> Cerrar Sesión
        </button>
      </div>

      <div className='usuarioLogueado'>
          <h2 className='txtUsuario'>Usuario: {UsuarioL}</h2>
      </div>

      <div className="input-section">
        <h2>Entrada:</h2>
        <textarea id="inputArea" value={TextoEntrada} onChange={(e) => SetTextoEntrada(e.target.value)}/>
      </div>

      <div className="output-section">
        <h2>Salida:</h2>
        <textarea id="outputArea" placeholder="Salida" value={TextoSalida} readOnly/>
      </div>
    </div>
  );
}

export default App;