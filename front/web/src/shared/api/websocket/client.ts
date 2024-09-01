export class WebSocketCLient {
  private ws: WebSocket;

  constructor(private url: string, private onMessageHandler: <T>(data: T) => void) {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log('Connected to: ', this.url);
    }
    this.ws.onmessage = (data) => {
      this.onMessageHandler(data);
    };
    this.ws.onclose = () => {
      console.log("disconnected");
    };
    this.ws.onerror = (error) => {
      console.log("error", error);
    };
  }

  public sendMessage(msg: string) {
    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(msg);
    } else {
      console.log('WebSocket is not opened!')
    }
  }
}