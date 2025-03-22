CREATE TABLE IF NOT EXISTS `orders_items` (
    `id` INT UNSIGNED AUTO_INCREMENT NOT NULL,
    `orderId` INT UNSIGNED NOT NULL,
    `productId` INT UNSIGNED NOT NULL,
    `quantity` INT UNSIGNED NOT NULL,
    `price` DECIMAL(10,2) NOT NULL,
    
    
    PRIMARY KEY (`id`),
    FOREIGN KEY (`orderId`) REFERENCES orders(`id`),
    FOREIGN KEY (`productId`) REFERENCES products(`id`)
);