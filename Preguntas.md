# Apartado 1

## ¿Qué pasos tiene que dar un proceso emisor para enviar un mensaje? ¿Y un receptor para recibirlo?
En el primer paso, el cliente que desea establecer la conexión envía al servidor un paquete SYN o segmento SYN (del inglés synchronize = “sincronizar”) con un número de secuencia individual y aleatorio. Este número garantiza la transmisión completa en el orden correcto (sin duplicados).

Si el servidor ha recibido el segmento, confirma el establecimiento de la conexión mediante el envío de un paquete SYN-ACK (del inglés acknowledgement = “confirmación”) incluido el número de secuencia del cliente después de sumarle 1. De forma adicional, transmite un número de secuencia propio al cliente.

Para finalizar, el cliente confirma la recepción del segmento SYN-ACK mediante el envío de un paquete ACK propio, que en este caso cuenta con el número de secuencia del servidor después de sumarle 1. En este punto también puede transmitir ya los primeros datos al servidor.

-- 

(1). Crear un conector (socket)

(2). Enlace de dirección IP y número de puerto del lado del servidor

(3). Utilice el método Listen () para abrir el lado del servidor de supervisión

(4). Utilice el método Accept () para intentar establecer una conexión con el lado cliente-servidor

(5). Utilice el método Connect () para establecer una conexión con el servidor-cliente

(5). Utilice el método Send () para enviar un mensaje al host conectado

(6). Utilice el método Recive () para recibir mensajes del host que estableció la conexión (conexión confiable)

--

## ¿Qué es big endian o little endian? 
Sistema de codificación de los bits dentro de un byte que determina el orden 
en el que se encuentran

Cadena de Bytes: FF FA
Big endian: FF FA
Little endian: FA FF


## ¿Qué son RTT, latencia, ventana de transmisión y el ancho de banda? 
RTT: Se llama RTT (round-trip time) al tiempo que tarda un paquete de datos enviado desde un emisor en volver a este mismo emisor habiendo pasado por el receptor de destino.

Latencia: Latencia de red es la suma de retardos temporales dentro de una red. Un retardo es producido por la demora en la propagación y transmisión de paquetes dentro de la red.

Ventana de Transmisión: Con la ventana de transmision se permite al emisor transmitir múltiples segmentos de información antes de comenzar la espera para que el receptor le confirme la recepción de los segmentos.

Ancho de Banda: El ancho de banda se mide como la cantidad de datos que se pueden transferir entre dos puntos de una red en un tiempo específico.


## ¿Cuánto tiempo cuesta transmitir un mensaje en TCP / IP? ¿De qué factores depende?

Lo que cueste el RTT

La velocidad de transmisión de datos que puede soportar TCP está limitada por varios factores.
 Algunos de estos son:
 	-Round-Trip Time (RTT)
 	-Velocidad de enlace más baja de las rutas de red involucradas
 	-Frecuencia de pérdida de paquetes
 	-La velocidad a la que los nuevos datos pueden estar disponibles para su transmisión.
 	-El tamaño máximo posible de la ventana de recepción de TCP.



# Apartado 2

## Cuál es la mejor manera de codificar esa respuesta? 
Como una cadena de bytes debidamente ordenada. Que haya un consenso entre emisor y receptor de como los bytes van ha estar ordenados.

## ¿Qué sucedería si el cliente se ejecuta en una máquina big endian y el servidor en una máquina little endian?
Que leerían los bytes de forma incorrecta y realizaría acciones erróneamente
Para que todo fuese correcto, se debería usar algún sistema de reorganización, como el uso
de la cadena BOM (Byte Order Mark).