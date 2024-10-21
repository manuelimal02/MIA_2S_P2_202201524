package main

import (
	"Proyecto1/Analizador"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HabilitarCors(respuesta *http.ResponseWriter) {
	(*respuesta).Header().Set("Access-Control-Allow-Origin", "*")
	(*respuesta).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*respuesta).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func AnalizarEntrada(respuesta http.ResponseWriter, solicitud *http.Request) {
	HabilitarCors(&respuesta)
	if solicitud.Method == http.MethodPost {
		body, err := ioutil.ReadAll(solicitud.Body)
		if err != nil {
			http.Error(respuesta, "Error En La Solicitud.", http.StatusInternalServerError)
			return
		}
		result := Analizador.Analizar(string(body))
		fmt.Fprintf(respuesta, result)
		return
	}
	http.Error(respuesta, "Método No permitido", http.StatusMethodNotAllowed)
}

func main() {
	http.HandleFunc("/AnalizadorGo/ProcesarComando", AnalizarEntrada)
	fmt.Println("-------------------------------------------")
	fmt.Println("Servidor corriendo en localhost:8080")
	fmt.Println("-------------------------------------------")
	http.ListenAndServe(":8080", nil)
}
