package com.o11y.poc.sales

data class AddProductPricingDto(
    val productId: String,
    val price: Double,
    val discount: Double?
)