import React, { useState } from "react";
import Swal from "sweetalert2";
import "./Visualizador.css";

const Visualizador = () => {
  const [rutas, setRutas] = useState([]); 
  const [partitions, setPartitions] = useState([]);
  const [selectedDisk, setSelectedDisk] = useState("");

  // Función para obtener las rutas de los discos
  const handleGetPathDisk = async () => {
    try {
      const response = await fetch('http://localhost:8080/AnalizadorGo/ObtenerDiscos', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: "obtenerdiscosruta",
      });

      const text = await response.text();

      if (text.includes("No hay particiones montadas.")) {
        Swal.fire(text, "No hay particiones montadas", "error");
      } else {
        const rutasObtenidas = text.split('\n').filter(ruta => ruta.trim() !== "");
        setRutas(rutasObtenidas);
        Swal.fire("Mostrando Discos", "Particiones Montadas", "success");
      }

    } catch (error) {
      Swal.fire("Error Al Enviar Texto: "+error, "Error Desconocido", "error");
    }
  };

  // Función para obtener las particiones de un disco específico
  const fetchPartitions = (diskPath) => {
    setSelectedDisk(getDiskName(diskPath)); // Guardar el disco seleccionado

    // Hacer la solicitud al backend para obtener las particiones
    fetch("http://localhost:8080/AnalizadorGo/ObtenerParticiones", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ path: diskPath }), 
    })
      .then((response) => response.json())
      .then((data) => {
        console.log("Particiones obtenidas:", data);
        setPartitions(data || []); 
      })
      .catch((error) => {
        console.error("Error al obtener particiones:", error);
        setPartitions([]); 
      });
  };

  // Función para obtener el nombre del disco a partir de la ruta (por ejemplo: "disco1.mia")
  const getDiskName = (diskPath) => {
    const parts = diskPath.split('/');
    return parts[parts.length - 1]; // Devolver la última parte de la ruta
  };

  return (
    <div className="unique-container">
      <button id="logoutBtn" className="unique-btn" onClick={handleGetPathDisk}>
        <i className="fas fa-play"></i> Mostrar Discos
      </button>

      <h2>Discos Creados</h2>
      <div className="unique-disks">
        {rutas.length > 0 ? (
          rutas.map((disk, index) => (
            <div key={index} className="unique-disk-item" onClick={() => fetchPartitions(disk)}>
              <i className="fas fa-hdd unique-disk-icon"></i> 
              <h5>Disco: {getDiskName(disk)}</h5>
            </div>
          ))
        ) : (
          <p>No se han creado discos aún.</p>
        )}
      </div>

      {selectedDisk && (
        <div className="unique-partitions">
          <h3>Particiones del Disco: {selectedDisk}</h3>
          {partitions && partitions.length > 0 ? (
            <ul className="unique-partition-list">
              {partitions.map((partition, index) => (
                <li key={index} className="unique-partition-item">
                  <strong>Nombre:</strong> {partition.name} | <strong>Tipo:</strong> {partition.type} |{" "}
                  <strong>Tamaño:</strong> {partition.size} | <strong>Inicio:</strong> {partition.start}
                </li>
              ))}
            </ul>
          ) : (
            <p>No se encontraron particiones para este disco.</p>
          )}
        </div>
      )}
    </div>
  );
};

export default Visualizador;
