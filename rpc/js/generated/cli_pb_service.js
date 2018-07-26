// package: rpc
// file: cli.proto

var cli_pb = require("./cli_pb");
var grpc = require("grpc-web-client").grpc;

var Cli = (function () {
  function Cli() {}
  Cli.serviceName = "rpc.Cli";
  return Cli;
}());

Cli.PublishTx = {
  methodName: "PublishTx",
  service: Cli,
  requestStream: false,
  responseStream: false,
  requestType: cli_pb.Transaction,
  responseType: cli_pb.Response
};

Cli.GetTransaction = {
  methodName: "GetTransaction",
  service: Cli,
  requestStream: false,
  responseStream: false,
  requestType: cli_pb.TransactionKey,
  responseType: cli_pb.Transaction
};

Cli.GetBalance = {
  methodName: "GetBalance",
  service: Cli,
  requestStream: false,
  responseStream: false,
  requestType: cli_pb.Key,
  responseType: cli_pb.Value
};

Cli.GetState = {
  methodName: "GetState",
  service: Cli,
  requestStream: false,
  responseStream: false,
  requestType: cli_pb.Key,
  responseType: cli_pb.Value
};

Cli.GetBlock = {
  methodName: "GetBlock",
  service: Cli,
  requestStream: false,
  responseStream: false,
  requestType: cli_pb.BlockKey,
  responseType: cli_pb.BlockInfo
};

exports.Cli = Cli;

function CliClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

CliClient.prototype.publishTx = function publishTx(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  grpc.unary(Cli.PublishTx, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          callback(Object.assign(new Error(response.statusMessage), { code: response.status, metadata: response.trailers }), null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
};

CliClient.prototype.getTransaction = function getTransaction(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  grpc.unary(Cli.GetTransaction, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          callback(Object.assign(new Error(response.statusMessage), { code: response.status, metadata: response.trailers }), null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
};

CliClient.prototype.getBalance = function getBalance(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  grpc.unary(Cli.GetBalance, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          callback(Object.assign(new Error(response.statusMessage), { code: response.status, metadata: response.trailers }), null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
};

CliClient.prototype.getState = function getState(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  grpc.unary(Cli.GetState, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          callback(Object.assign(new Error(response.statusMessage), { code: response.status, metadata: response.trailers }), null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
};

CliClient.prototype.getBlock = function getBlock(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  grpc.unary(Cli.GetBlock, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          callback(Object.assign(new Error(response.statusMessage), { code: response.status, metadata: response.trailers }), null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
};

exports.CliClient = CliClient;

