import './App.css';
import '@fortawesome/fontawesome-free/css/all.min.css';
import React, { useState } from 'react';

function App() {
  const [inputText, setInputText] = useState('');
  const [outputText, setOutputText] = useState('');

  const handleFileSelect = (event) => {
    const fileInput = event.target.files[0];
    if (fileInput && fileInput.name.endsWith('.smia')) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setInputText(e.target.result);
      };
      reader.readAsText(fileInput);
    } else {
      alert("Se debe seleccionar un archivo con la extensión '.smia'");
    }
  };

  const triggerFileSelect = () => {
    const fileInput = document.createElement('input');
    fileInput.type = 'file';
    fileInput.accept = '.smia';
    fileInput.onchange = handleFileSelect;
    fileInput.click();
  };

  const handleExecute = async () => {
    try {
      const response = await fetch('http://localhost:8080/AnalizadorGo/ProcesarComando', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: inputText,
      });
      const text = await response.text();
      setOutputText(text);
    } catch (error) {
      console.error('Error al enviar el texto:', error);
      setOutputText('Error al procesar los comandos de entrada.'); 
    }
  };

  return (
    <div className="container">
      <h1>Sistema De Archivos EXT2</h1>

      <div className="buttons">
        <button id="selectBtn" onClick={triggerFileSelect}>
          <i className="fas fa-folder-open"></i> Seleccionar
        </button>
        <button id="executeBtn" onClick={handleExecute}>
          <i className="fas fa-play"></i> Ejecutar
        </button>
      </div>

      <div className="input-section">
        <h2>Entrada:</h2>
        <textarea id="inputArea" value={inputText} onChange={(e) => setInputText(e.target.value)}/>
      </div>

      <div className="output-section">
        <h2>Salida:</h2>
        <textarea id="outputArea" placeholder="Salida" value={outputText} readOnly/>
      </div>
    </div>
  );
}

export default App;