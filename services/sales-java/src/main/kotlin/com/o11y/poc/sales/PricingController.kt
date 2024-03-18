package com.o11y.poc.sales

import org.slf4j.LoggerFactory
import org.springframework.web.bind.annotation.*


@RestController
@RequestMapping("/pricing")
class PricingController() {

    companion object {
        private val logger = LoggerFactory.getLogger(PricingController::class.java)
        private val prices = mutableListOf(
            ProductPricingInfo("1", 100.0, 0.0),
            ProductPricingInfo("2", 200.0, 0.0),
            ProductPricingInfo("3", 300.0, 0.0),
            ProductPricingInfo("4", 400.0, 0.1),
            ProductPricingInfo("5", 500.0, 0.15),
        )
    }

    @GetMapping("")
    fun getPrices(
    ): List<ProductPricingInfo> {
        return prices
    }

    @GetMapping("/{productIds}")
    fun getProductsPrices(
        @PathVariable productIds: List<String>
    ): List<ProductPricingInfo> {
        return prices.filter { it.productId in productIds }
    }

    @PostMapping("/")
    fun addSale(@RequestBody pricing: AddProductPricingDto): ProductPricingInfo {
        val price = ProductPricingInfo(pricing.productId, pricing.price, pricing.discount ?: 0.0)
        prices += price
        logger.info("Added new price: $price")
        return price
    }
}
