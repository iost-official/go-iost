// package: rpc
// file: cli.proto

import * as cli_pb from "./cli_pb";
import {grpc} from "grpc-web-client";

type CliPublishTx = {
  readonly methodName: string;
  readonly service: typeof Cli;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof cli_pb.Transaction;
  readonly responseType: typeof cli_pb.Response;
};

type CliGetTransaction = {
  readonly methodName: string;
  readonly service: typeof Cli;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof cli_pb.TransactionKey;
  readonly responseType: typeof cli_pb.Transaction;
};

type CliGetBalance = {
  readonly methodName: string;
  readonly service: typeof Cli;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof cli_pb.Key;
  readonly responseType: typeof cli_pb.Value;
};

type CliGetState = {
  readonly methodName: string;
  readonly service: typeof Cli;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof cli_pb.Key;
  readonly responseType: typeof cli_pb.Value;
};

type CliGetBlock = {
  readonly methodName: string;
  readonly service: typeof Cli;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof cli_pb.BlockKey;
  readonly responseType: typeof cli_pb.BlockInfo;
};

export class Cli {
  static readonly serviceName: string;
  static readonly PublishTx: CliPublishTx;
  static readonly GetTransaction: CliGetTransaction;
  static readonly GetBalance: CliGetBalance;
  static readonly GetState: CliGetState;
  static readonly GetBlock: CliGetBlock;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }
export type ServiceClientOptions = { transport: grpc.TransportConstructor }

interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: () => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}

export class CliClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: ServiceClientOptions);
  publishTx(
    requestMessage: cli_pb.Transaction,
    metadata: grpc.Metadata,
    callback: (error: ServiceError, responseMessage: cli_pb.Response|null) => void
  ): void;
  publishTx(
    requestMessage: cli_pb.Transaction,
    callback: (error: ServiceError, responseMessage: cli_pb.Response|null) => void
  ): void;
  getTransaction(
    requestMessage: cli_pb.TransactionKey,
    metadata: grpc.Metadata,
    callback: (error: ServiceError, responseMessage: cli_pb.Transaction|null) => void
  ): void;
  getTransaction(
    requestMessage: cli_pb.TransactionKey,
    callback: (error: ServiceError, responseMessage: cli_pb.Transaction|null) => void
  ): void;
  getBalance(
    requestMessage: cli_pb.Key,
    metadata: grpc.Metadata,
    callback: (error: ServiceError, responseMessage: cli_pb.Value|null) => void
  ): void;
  getBalance(
    requestMessage: cli_pb.Key,
    callback: (error: ServiceError, responseMessage: cli_pb.Value|null) => void
  ): void;
  getState(
    requestMessage: cli_pb.Key,
    metadata: grpc.Metadata,
    callback: (error: ServiceError, responseMessage: cli_pb.Value|null) => void
  ): void;
  getState(
    requestMessage: cli_pb.Key,
    callback: (error: ServiceError, responseMessage: cli_pb.Value|null) => void
  ): void;
  getBlock(
    requestMessage: cli_pb.BlockKey,
    metadata: grpc.Metadata,
    callback: (error: ServiceError, responseMessage: cli_pb.BlockInfo|null) => void
  ): void;
  getBlock(
    requestMessage: cli_pb.BlockKey,
    callback: (error: ServiceError, responseMessage: cli_pb.BlockInfo|null) => void
  ): void;
}

