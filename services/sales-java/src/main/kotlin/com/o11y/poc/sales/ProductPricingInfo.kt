package com.o11y.poc.sales


data class ProductPricingInfo(
    val productId: String,
    val price: Double,
    val discount: Double
) {
    val total: Double
        get() = price * (1 - discount)
}