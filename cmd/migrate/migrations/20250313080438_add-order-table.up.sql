CREATE TABLE IF NOT EXISTS orders (
    `id` INT UNSIGNED AUTO_INCREMENT NOT NULL,
    `userId` INT UNSIGNED NOT NULL,
    `total` INT UNSIGNED NOT NULL,
    `status` ENUM('pending', 'completed', 'cancelled') NOT NULL DEFAULT 'pending',
    `address` TEXT NOT NULL,
    `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY(`id`),
    FOREIGN KEY(`userId`) REFERENCES users(`id`)
)