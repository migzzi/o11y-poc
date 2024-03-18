package com.o11y.poc.sales

import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator
import io.opentelemetry.context.propagation.ContextPropagators
import io.opentelemetry.exporter.otlp.http.logs.OtlpHttpLogRecordExporter
import io.opentelemetry.exporter.otlp.http.metrics.OtlpHttpMetricExporter
import io.opentelemetry.exporter.otlp.http.trace.OtlpHttpSpanExporter
import io.opentelemetry.sdk.OpenTelemetrySdk
import io.opentelemetry.sdk.logs.SdkLoggerProvider
import io.opentelemetry.sdk.logs.export.BatchLogRecordProcessor
import io.opentelemetry.sdk.metrics.SdkMeterProvider
import io.opentelemetry.sdk.metrics.export.PeriodicMetricReader
import io.opentelemetry.sdk.resources.Resource
import io.opentelemetry.sdk.trace.SdkTracerProvider
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor
import io.opentelemetry.semconv.ResourceAttributes
import org.springframework.beans.factory.annotation.Value
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration


@Configuration
class OTELConfig {

    @Value("\${otel.exporter.otlp.endpoint}")
    private lateinit var otlpEndpoint: String

    @Value("\${otel.service.name:sales-service}")
    private lateinit var serviceName: String

    @Value("\${otel.service.version:0.1.0}")
    private lateinit var serviceVersion: String


//    @Bean
//    fun propagator(): TextMapPropagator {
//        return io.opentelemetry.extension.trace.propagation.Pr.getInstance()
//    }
//

    @Bean
    fun openTelemetry(): OpenTelemetry {
        val resource: Resource = Resource.getDefault().toBuilder()
            .put(ResourceAttributes.SERVICE_NAME, serviceName)
            .put(ResourceAttributes.SERVICE_VERSION, serviceVersion).build()
        val sdkTracerProvider = SdkTracerProvider.builder()
            .addSpanProcessor(BatchSpanProcessor.builder(
                OtlpHttpSpanExporter.builder().setEndpoint("$otlpEndpoint/v1/traces").build()
            ).build())
            .setResource(resource)
            .build()
        val sdkMeterProvider = SdkMeterProvider.builder()
            .registerMetricReader(PeriodicMetricReader.builder(
                OtlpHttpMetricExporter.builder().setEndpoint("$otlpEndpoint/v1/metrics").build()
            ).build())
            .setResource(resource)
            .build()
        val sdkLoggerProvider = SdkLoggerProvider.builder()
            .addLogRecordProcessor(
                BatchLogRecordProcessor.builder(
                    OtlpHttpLogRecordExporter.builder().setEndpoint("$otlpEndpoint/v1/logs").build()
                ).build()
            )
            .setResource(resource)
            .build()
        return OpenTelemetrySdk.builder()
            .setTracerProvider(sdkTracerProvider)
            .setMeterProvider(sdkMeterProvider)
            .setLoggerProvider(sdkLoggerProvider)
            .setPropagators(ContextPropagators.create(W3CTraceContextPropagator.getInstance()))
            .buildAndRegisterGlobal()
    }


}