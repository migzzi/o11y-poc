import { diag, DiagLogLevel, DiagConsoleLogger } from "@opentelemetry/api";
import { NodeSDK } from "@opentelemetry/sdk-node";
import { getNodeAutoInstrumentations } from "@opentelemetry/auto-instrumentations-node";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { OTLPMetricExporter } from "@opentelemetry/exporter-metrics-otlp-http";
// import { PeriodicExportingMetricReader } from "@opentelemetry/sdk-metrics";
import { ConsoleSpanExporter } from "@opentelemetry/sdk-trace-node";

diag.setLogger(new DiagConsoleLogger(), DiagLogLevel.INFO);

console.log(`OTEL_COLLECTOR_URL: ${process.env.OTEL_COLLECTOR_URL}`);

const otlpTracesExporter = new OTLPTraceExporter({
  url: `${process.env.OTEL_COLLECTOR_URL}/v1/traces`,
});
// const consoleExporter = new ConsoleSpanExporter();
// const otlpMetricExporter = new OTLPMetricExporter({
//   url: `${process.env.OTEL_COLLECTOR_URL}/v1/metrics`,
// });

const sdk = new NodeSDK({
  traceExporter: otlpTracesExporter,
  //   metricReader: new PeriodicExportingMetricReader({
  //     exporter: otlpMetricExporter,
  //     exportIntervalMillis: 10000,
  //   }),
  instrumentations: [getNodeAutoInstrumentations()],
});

try {
  sdk.start();
  console.log("Tracing initialized");
} catch (err) {
  console.log("Error initializing tracing", err);
}
