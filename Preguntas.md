# Apartado 1

## ¿Qué pasos tiene que dar un proceso emisor para enviar un mensaje?
-Leer Florian Westphal en Fedora Magazine: TCP window scaling, timestamps and SACK


## ¿Y un receptor para recibirlo?
-Same as before


## ¿Qué es big endian o little endian? 
Sistema de codificación de los bits dentro de un byte que determina el orden 
en el que se encuentran

Cadena de Bytes: FF FA
Big endian: FF FA
Little endian: FA FF


## ¿Qué son RTT, latencia, ventana de transmisión y el ancho de banda? 
RTT:
Latencia:
Ventana de Transmisión:
Ancho de Banda:


## ¿Cuánto tiempo cuesta transmitir un mensaje en TCP / IP? ¿De qué factores depende?
-MIRAR COSTE EN DIAPOS-
Depende de los factores anteriormente descritos


# Apartado 2

## Cuál es la mejor manera de codificar esa respuesta? 
Como una cadena de bytes debidamente ordenada.

##¿Qué sucedería si el cliente se ejecuta en una máquina big endian y el servidor en una máquina little endian?
Que leerían los bytes de forma incorrecta y realizaría acciones erróneamente
Para que todo fuese correcto, se debería usar algún sistema de reorganización, como el uso
de la cadena BOM (Byte Order Mark).