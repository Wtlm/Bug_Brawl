let socket = null;
let messageHandler = null;

export const getSocket = () => {
  if (!socket || socket.readyState === WebSocket.CLOSED) {
    socket = new WebSocket(`${location.origin.replace(/^http/, "ws")}/ws`);

    socket.onopen = () => {
      console.log("WebSocket connected");
    };

    socket.onmessage = (event) => {
      console.log("Received message:", event.data);
      try {
        const data = JSON.parse(event.data);
        if (data.error) {
          alert(data.error);
          socket?.close();
          return;
        }

        messageHandler?.(data);
      } catch (error) {
        console.error("Error parsing message:", error);
      }
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
      alert("Connection error. Please try again.");
    };

    socket.onclose = (event) => {
      console.log("WebSocket closed:", event.code, event.reason);
      socket = null;
    };
  }
  return socket;
};

export function setMessageHandler(handler) {
  messageHandler = handler;
}

export const closeSocket = () => {
  if (socket) {
    socket.close();
    socket = null;
    messageHandler = null;
  }
};