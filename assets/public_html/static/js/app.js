var host = window.location.href.split("/")[2];
var ws;
var app = new Vue({
  el: '#app',
  data: {
    loading: false,
    type: null,
    pendingAct: [],
    paladins: [
      { id: "0075", name: 'PEREGRINE PALADIN / LARUT', rawhp: 1000, modhp: 0, winloss: "-/-", symbol: "", inprice: "0", color: "green" },
      { id: "0256", name: 'MAGE PALADIN / DISTRIER', rawhp: 1000, modhp: 0, winloss: "-/-", symbol: "", inprice: "0", color: "green" },
      { id: "0295", name: 'FLASH PALADIN / IBERT', rawhp: 1000, modhp: 0, winloss: "-/-", symbol: "", inprice: "0", color: "green" },
      { id: "0553", name: 'LADY PALADIN / MIRELIA', rawhp: 1000, modhp: 0, winloss: "-/-", symbol: "", inprice: "0", color: "green" },
      { id: "0880", name: 'PALADIN OF TRUTH / INZAGHI', rawhp: 1000, modhp: 0, winloss: "-/-", symbol: "", inprice: "0", color: "green" }
    ],
    isConnected: false
  },
  computed: {
    activePaladins: function() {
      var arr = [];
      for (var i in this.paladins) {
        if (this.paladins[i].type == this.type) {
          arr.push(this.paladins[i]);
        }
      }
      return arr;
    },
  },
  methods: {
    setType: function (type) {
      this.type = type;
    },
    formatSymbol: function (symbol) {
      return symbol ? symbol.substring(0, symbol.length - 3) : symbol;
    },
    formatDuration: function (milis) {
      var day = parseInt(milis / 86400000);
      var hour = parseInt((milis - day * 86400000) / 3600000);
      var min = parseInt((milis - day * 86400000 - hour * 3600000) / 60000);
      return day + "d " + hour + "h " + min + "m";
    },
    duration: function(fromMilis) {
      return this.formatDuration(new Date().getTime() - fromMilis);
    },
    action: function (id, act) {
      app.loading = true;
      var request = $.ajax({
        url: "http://" + host + "/regus/action/info",
        type: 'GET',
        data: { id: id, act : act, symbol: $("#action-symbol-" + id).val() } ,
        contentType: 'application/json; charset=utf-8'
      });
      
      request.done(function(data) {
        app.pendingAct = [id, act];
        app.loading = false;
        $('#modal-primary-body').html(data.msg);
        $('#modal-primary').modal();
      });

      request.fail(function(jqXHR, textStatus) {
        app.loading = false;
        $('#modal-primary-body').html(jqXHR.responseText);
        $('#modal-primary').modal();
      });
    },
    processAction: function () {
      if (app.pendingAct.length != 2) {
        $('#modal-primary').modal('hide');
        return;
      }
      var id = app.pendingAct[0];
      var act = app.pendingAct[1];
      app.pendingAct = [];
      app.loading = true;
      var request = $.ajax({
        url: "http://" + host + "/regus/action",
        type: 'GET',
        data: { id: id, act : act, symbol: $("#action-symbol-" + id).val() } ,
        contentType: 'application/json; charset=utf-8'
      });
      
      request.done(function(data) {
        app.paladins = data;
        app.loading = false;
        $('#modal-primary').modal('hide');
      });
      
      request.fail(function(jqXHR, textStatus) {
        app.loading = false;
        $('#modal-primary-body').html(jqXHR.responseText);
      });
    }
  }
});

function initWs() {
  ws = new WebSocket("ws://" + host + "/regus/ws");
  ws.onopen = function(evt) {
    app.isConnected = true;
  }
  ws.onclose = function(evt) {
    app.isConnected = false;
    ws = null;
  }
  ws.onmessage = function(evt) {
    var data = JSON.parse(evt.data);
    var prices = data.prices;
    var paladins = data.full? data.full : Object.assign({}, app.paladins);
    for (i in prices) {
      for (p in paladins) {
        if (paladins[p].id == i) {
          paladins[p].curprice = prices[i];
          if (paladins[p].inprice) {
            paladins[p].change = (100 * (prices[i] - paladins[p].inprice) / paladins[p].inprice).toFixed(2);
            if (paladins[p].change >= 0) paladins[p].color = "aqua";
            else if (paladins[p].change < 0) paladins[p].color = "red";
          } else {
            paladins[p].change = 0;
            paladins[p].color = "green";
          }
        }
      }
    }
    app.paladins = paladins
  }
  ws.onerror = function(evt) {
    console.log("WS ERROR: " + evt.data);
    app.isConnected = false;
  }
}

$(document).ready(function () {
  initWs();

  // app.loading = true;
  // $.get("http://" + host + "/regus/paladins", function(data) {
  //   app.paladins = data;
  //   initWs();
  //   app.loading = false;
  // });
});
