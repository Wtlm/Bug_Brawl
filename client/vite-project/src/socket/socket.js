let socket = null;

export const getSocket = () => {
  if (!socket || socket.readyState === WebSocket.CLOSED) {
    socket = new WebSocket(`${location.origin.replace(/^http/, "ws")}/ws`);
  }
  return socket;
};

export const closeSocket = () => {
  if (socket) {
    socket.close();
    socket = null;
  }
};