(function(win){
    var client = function (options) {
        this.options = options || {};
        this.ws;
        this.connect(10, 1000*10);
        
    }

    client.prototype.connect = function(maxConnectTime, delay) {
        var self = this

        var domain = "ws://192.168.0.87:10000/im"
        // var domain = "ws://localhost:10000/im"
        if ("WebSocket" in window) {
            console.log("WebSocket")
            self.ws = new window.WebSocket(domain)
        } else if ("MozWebSocket" in window) {
            console.log("MozWebSocket")
            self.ws = new window.MozWebSocket(domain)
        } else {
            alert("不支持websocket")
            return
        }
        
        
        var flag = false;
        var heartbeatInterval
        self.ws.onopen = function(ev) {
            auth()
        }
        self.ws.onmessage = function (ev) {
            receiver = JSON.parse(ev.data)
            if (receiver.op == 3) {
                flag = true
                heartbeatInterval = setInterval(heartbeat, 30*1000)
            }
            if (!flag) {
                setTimeout(auth, 10 * 1000)
            }

            flag = true;
            if (self.options.notify && flag) {
                self.options.notify(ev.data)
            }
        }

        self.ws.onerror = function(ev) {
            console.log(ev);
        }
        self.ws.onclose = function(ev) {
            if (heartbeatInterval) {
                clearInterval(heartbeatInterval)
            }
            setTimeout(reconnect, delay)
        }
        function auth() {
            self.ws.send(JSON.stringify({
                ver:"33333",
                sessionId:"xxxxx",
                op:2,
                body:"我是来认证的2"
            }))
        }
        function reconnect() {
            self.connect(maxConnectTime--, delay*1.5)
        }
        function heartbeat() {
            self.ws.send(JSON.stringify({
                ver:"1",
                op:0,
            }))
        }
        
    }

    client.prototype.send = function(msg) {
        console.log("send msg",msg)
        this.ws.send(JSON.stringify({
            ver:"2",
            op:4,
            body:msg,
        }))
    }
    win["client"] = client;
   
})(window);