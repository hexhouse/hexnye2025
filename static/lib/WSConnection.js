export default class WSConnection {
  constructor(url) {
    this.url = url;
  }

  connect(opts = {}) {
    const { binaryType } = opts;
    if (this.ws)
      this.ws.close();
    this.ws = new WebSocket(this.url);
    this.ws.binaryType = binaryType || "arraybuffer";
    this.ws.onclose = e => {
      this.open = false;
      setTimeout(() => this.connect(opts), 1000);
      this.onclose && this.onclose();
    };
    this.ws.onopen = e => {
      this.open = true;
    };
    this.ws.onmessage = e => {
      this.onmessage((e.data instanceof ArrayBuffer || e.data instanceof Blob) ? e.data : JSON.parse(e.data));
    };
  }

  send(obj) {
    return this.ws.send(ArrayBuffer.isView(obj) ? obj : JSON.stringify(obj));
  }
}


