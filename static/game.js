/*
 * Enhanced Balloon Game
 * Based on original by Kamesh Sampath
 *
 * Features added:
 * - Significantly faster balloons with progressive difficulty
 * - Negative point balloons (with special appearance)
 */

class BalloonGame {
    constructor(character, playerName) {
        console.log("Initializing game for:", playerName, "as", character);
        this.character = character;
        this.playerName = playerName;
        this.score = 0;
        this.bonusHits = 0;
        this.regularHits = 0;
        this.negativeHits = 0;
        this.isActive = true;
        this.level = 1;
        this.levelUpTime = 20000; // Level up every 20 seconds (faster progression)
        this.gameStartTime = Date.now();

        // Canvas setup
        this.canvas = document.getElementById("gameCanvas");
        this.ctx = this.canvas.getContext("2d");
        this.resizeCanvas();

        // Game state
        this.balloons = [];
        this.popEffects = [];
        this.lastSpawnTime = 0;
        this.spawnInterval = 2000;
        this.balloonRadius = 30;
        this.hitAreaMultiplier = 1.5;

        // Enhanced speed configurations
        this.baseSpeed = 2.0; // Starting with faster base speed (was 1.0)
        this.speedMultiplier = 1.0;
        this.maxSpeedMultiplier = 4.0; // Higher maximum speed (was 3.0)
        this.speedIncreaseRate = 0.3; // Faster speed progression per level (was 0.2)

        // Spawn rate configuration
        this.minSpawnInterval = 300; // Minimum spawn interval in ms (faster)
        this.spawnRateDecreasePerLevel = 250; // Faster balloon spawn rate per level

        // Negative balloon configuration
        this.negativeColor = "black";
        this.negativeProbability = 0.15; // 15% chance of negative balloon

        // Initialize WebSocket
        this.connectWebSocket();

        // Set up event listeners
        window.addEventListener("resize", () => this.resizeCanvas());
        this.canvas.addEventListener("click", (e) => this.handleClick(e));

        // Game Config
        fetch("/config")
            .then((response) => response.json())
            .then((config) => {
                this.gameConfig = config;
                console.log("Loaded game config:", config);

                // Set negative color to one not in any favorite colors
                this.setNegativeColor();

                // Start game loop
                this.gameLoop();

                // Start level timer
                this.startLevelTimer();
            })
            .catch((error) =>
                console.error("Failed to load game config:", error)
            );
    }

    // Set negative colors to be EXACTLY the same as favorite colors
    setNegativeColor() {
        if (!this.gameConfig || !this.gameConfig.colors) return;

        // Get current character's favorite colors - use these EXACT colors for negative balloons
        this.negativeColors = this.gameConfig.character_favorites[this.character] || [];

        console.log("Set negative balloon colors to match favorite colors:", this.negativeColors);

        // If we couldn't get favorite colors, fall back to black
        if (this.negativeColors.length === 0) {
            this.negativeColors = ["black"];
        }
    }

    // Start level timer to increase difficulty over time
    startLevelTimer() {
        setInterval(() => {
            if (this.isActive) {
                this.level++;

                // Apply progressive speed increase
                this.speedMultiplier = Math.min(
                    this.maxSpeedMultiplier,
                    1 + (this.level - 1) * this.speedIncreaseRate
                );

                // Make balloons spawn faster as level increases
                this.spawnInterval = Math.max(
                    this.minSpawnInterval,
                    2000 - (this.level - 1) * this.spawnRateDecreasePerLevel
                );

                console.log(`Level up! Level ${this.level}, Speed: ${this.speedMultiplier.toFixed(1)}x, Spawn: ${this.spawnInterval}ms`);

                // Show level up message
                const levelElement = document.getElementById("level");
                if (levelElement) {
                    levelElement.textContent = `Level: ${this.level}`;
                    levelElement.classList.add("level-up");
                    setTimeout(() => levelElement.classList.remove("level-up"), 1000);
                }

                // Increase difficulty for higher levels
                if (this.level > 3) {
                    // Increase negative balloon probability after level 3
                    this.negativeProbability = Math.min(0.25, 0.15 + (this.level - 3) * 0.02);
                }
            }
        }, this.levelUpTime);
    }

    resizeCanvas() {
        const container = this.canvas.parentElement;
        const rect = container.getBoundingClientRect();
        this.canvas.width = rect.width;
        this.canvas.height = 500;
        console.log(
            "Canvas resized to:",
            this.canvas.width,
            this.canvas.height
        );
    }

    connectWebSocket() {
        console.log("Connecting WebSocket for player:", this.playerName);
        this.ws = new WebSocket(
            `ws://${window.location.host}/ws/${this.playerName}`
        );

        this.ws.onopen = () => {
            console.log("WebSocket connected");
        };

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            console.log("Received WebSocket message:", data);
            if (data.type === "score_update") {
                this.updateScore(data.event);
            }
        };
    }

    createBalloon() {
        if (!this.gameConfig) return;

        // Get character's favorite colors from game config
        const favoriteColors = this.gameConfig.character_favorites[this.character];
        const bonusProbability = this.gameConfig.bonus_probability;

        // Available colors for regular balloons (excluding all similar negative colors)
        const regularColors = Object.keys(this.gameConfig.colors).filter(
            (color) => !favoriteColors.includes(color) && !this.negativeColors.includes(color)
        );

        // Determine if this will be a bonus, negative, or regular balloon
        const rand = Math.random();
        let isBonus = false;
        let isNegative = false;

        if (rand < this.negativeProbability) {
            isNegative = true;
        } else if (rand < this.negativeProbability + bonusProbability) {
            isBonus = true;
        }

        // Select color based on balloon type
        let color;
        if (isNegative) {
            // Select a random negative color from our similar colors array
            color = this.negativeColors[Math.floor(Math.random() * this.negativeColors.length)];
        } else if (isBonus) {
            color = favoriteColors[Math.floor(Math.random() * favoriteColors.length)];
        } else {
            color = regularColors[Math.floor(Math.random() * regularColors.length)];
        }

        // Calculate speed based on current level with more variation
        // Higher levels have higher potential speed variation
        const levelVariationBonus = Math.min(0.5, (this.level - 1) * 0.1); // Up to 0.5 bonus variation at level 6+
        const speedVariation = (Math.random() * (0.5 + levelVariationBonus)) + 0.75; // 0.75 to 1.25+ variation

        // Calculate base speed with all factors
        let speed = this.baseSpeed * this.speedMultiplier * speedVariation;

        // Special fast balloons at higher levels (10% chance after level 2)
        const isFastBalloon = !isNegative && this.level > 2 && Math.random() < 0.1;

        // Make negative balloons slower - more tempting to hit!
        if (isNegative) {
            // Negative balloons are 40-60% slower than normal balloons
            // They get relatively slower as levels increase, making them more tempting targets
            const slowFactor = 0.6 - Math.min(0.2, (this.level - 1) * 0.04);
            speed *= slowFactor;
        } else if (isFastBalloon) {
            // Fast balloons are 50% faster (only for non-negative balloons)
            speed *= 1.5;
        }

        const balloon = {
            x: Math.random() * (this.canvas.width - 60) + 30,
            y: this.canvas.height + 30,
            radius: this.balloonRadius,
            color: color,
            speed: speed,
            bobOffset: 0,
            bobSpeed: Math.random() * 0.05 + 0.02,
            bobTime: Math.random() * Math.PI * 2,
            scale: 1,
            isBonus: isBonus,
            isNegative: isNegative,
            isFast: isFastBalloon,
            sparkleAngle: 0,
            spikes: isNegative ? 5 : 0, // Spikes for negative balloons
        };

        this.balloons.push(balloon);
    }

    createPopEffect(x, y, color) {
        const particles = [];
        const particleCount = 8;

        for (let i = 0; i < particleCount; i++) {
            const angle = (i / particleCount) * Math.PI * 2;
            particles.push({
                x,
                y,
                color,
                speed: 5,
                angle,
                life: 1,
            });
        }

        this.popEffects.push({
            particles,
            age: 0,
        });
    }

    updateBalloons() {
        for (let i = this.balloons.length - 1; i >= 0; i--) {
            const balloon = this.balloons[i];
            balloon.y -= balloon.speed;
            balloon.bobTime += balloon.bobSpeed;
            balloon.bobOffset = Math.sin(balloon.bobTime) * 5;

            if (balloon.isBonus || balloon.isNegative) {
                balloon.sparkleAngle += 0.05;
            }

            if (balloon.y < -50) {
                this.balloons.splice(i, 1);
            }
        }

        // Update pop effects
        for (let i = this.popEffects.length - 1; i >= 0; i--) {
            const effect = this.popEffects[i];
            effect.age += 0.05;

            effect.particles.forEach((particle) => {
                particle.x += Math.cos(particle.angle) * particle.speed;
                particle.y += Math.sin(particle.angle) * particle.speed;
                particle.life -= 0.05;
            });

            if (effect.age > 1) {
                this.popEffects.splice(i, 1);
            }
        }
    }

    handleClick(event) {
        if (!this.isActive) return;

        const rect = this.canvas.getBoundingClientRect();
        const x = (event.clientX - rect.left) * (this.canvas.width / rect.width);
        const y = (event.clientY - rect.top) * (this.canvas.height / rect.height);

        for (let i = this.balloons.length - 1; i >= 0; i--) {
            const balloon = this.balloons[i];
            const dx = x - balloon.x;
            const dy = y - (balloon.y + balloon.bobOffset);
            const distance = Math.sqrt(dx * dx + dy * dy);

            if (distance < balloon.radius * this.hitAreaMultiplier) {
                // Create pop effect
                this.createPopEffect(
                    balloon.x,
                    balloon.y + balloon.bobOffset,
                    balloon.color
                );

                // Handle negative balloon locally with immediate feedback
                if (balloon.isNegative) {
                    this.handleNegativeBalloon();
                    // Update score locally
                    this.score = Math.max(0, this.score - 10);
                    const scoreElement = document.getElementById("score");
                    if (scoreElement) {
                        scoreElement.textContent = `Score: ${this.score}`;
                        scoreElement.classList.add("negative");
                        setTimeout(() => scoreElement.classList.remove("negative"), 1000);
                    }
                }

                // Send pop event to server
                if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                    this.ws.send(
                        JSON.stringify({
                            balloon_color: balloon.color,
                            event_ts: new Date().toISOString(),
                            player: this.playerName,
                            character: this.character,
                            negative_hit: balloon.isNegative
                        })
                    );
                }

                this.balloons.splice(i, 1);
                break;
            }
        }
    }

    handleNegativeBalloon() {
        this.negativeHits++;

        // Update UI
        const negativeElement = document.getElementById("negativeHits");
        if (negativeElement) {
            negativeElement.textContent = `Negative Hits: ${this.negativeHits}`;
            negativeElement.classList.add("negative");
            setTimeout(() => negativeElement.classList.remove("negative"), 1000);
        }
    }

    drawBalloon(balloon) {
        const ctx = this.ctx;
        const x = balloon.x;
        const y = balloon.y + balloon.bobOffset;

        // Draw string
        ctx.beginPath();
        ctx.moveTo(x, y + balloon.radius);
        ctx.lineTo(x, y + balloon.radius + 20);
        ctx.strokeStyle = "#666";
        ctx.lineWidth = 2;
        ctx.stroke();

        // Draw balloon
        if (balloon.isNegative) {
            // Draw deceptive negative balloon
            this.drawNegativeBalloon(balloon, x, y);

            // Add subtle downward arrow indicator for slower balloons
            this.drawSlowIndicator(balloon, x, y);
        } else {
            // Draw regular balloon
            ctx.beginPath();
            ctx.arc(x, y, balloon.radius, 0, Math.PI * 2);
            ctx.fillStyle = balloon.color;
            ctx.fill();

            // Add highlight to non-negative balloons
            const gradient = ctx.createRadialGradient(
                x - balloon.radius / 3,
                y - balloon.radius / 3,
                balloon.radius / 10,
                x,
                y,
                balloon.radius
            );
            gradient.addColorStop(0, "rgba(255, 255, 255, 0.4)");
            gradient.addColorStop(1, "rgba(255, 255, 255, 0)");
            ctx.fillStyle = gradient;
            ctx.fill();
        }

        // Add sparkle effect for bonus balloons
        if (balloon.isBonus) {
            // Draw sparkle
            const sparklePoints = 8;
            const outerRadius = balloon.radius * 1.3;
            const innerRadius = balloon.radius * 1.1;

            ctx.beginPath();
            for (let i = 0; i < sparklePoints * 2; i++) {
                const radius = i % 2 === 0 ? outerRadius : innerRadius;
                const angle = (i * Math.PI) / sparklePoints + balloon.sparkleAngle;
                const sparkleX = x + Math.cos(angle) * radius;
                const sparkleY = y + Math.sin(angle) * radius;

                if (i === 0) {
                    ctx.moveTo(sparkleX, sparkleY);
                } else {
                    ctx.lineTo(sparkleX, sparkleY);
                }
            }
            ctx.closePath();
            ctx.strokeStyle = "gold";
            ctx.lineWidth = 2;
            ctx.stroke();
        }

        // Add speed streaks behind fast balloons
        if (balloon.isFast) {
            // Draw speed streaks
            ctx.beginPath();
            for (let i = 1; i <= 3; i++) {
                const streakY = y + (i * 8);
                ctx.moveTo(x - 15, streakY);
                ctx.lineTo(x + 15, streakY);
            }
            ctx.strokeStyle = "rgba(255, 255, 255, 0.6)";
            ctx.lineWidth = 2;
            ctx.stroke();
        }
    }

    // Draw indicator for slow negative balloons
    drawSlowIndicator(balloon, x, y) {
        const ctx = this.ctx;

        // Very subtle downward-pointing arrow to hint at slower movement
        // Placed at the bottom of the balloon to be less obvious
        const arrowSize = 5;
        const arrowY = y + balloon.radius - 5;

        ctx.beginPath();
        ctx.moveTo(x, arrowY + arrowSize);
        ctx.lineTo(x - arrowSize, arrowY);
        ctx.lineTo(x + arrowSize, arrowY);
        ctx.closePath();

        // Use a very subtle color to make it challenging to notice
        ctx.fillStyle = "rgba(255, 0, 0, 0.3)";
        ctx.fill();
    }

    drawNegativeBalloon(balloon, x, y) {
        const ctx = this.ctx;

        // Draw balloon using EXACTLY the same appearance as regular balloons
        ctx.beginPath();
        ctx.arc(x, y, balloon.radius, 0, Math.PI * 2);
        ctx.fillStyle = balloon.color;
        ctx.fill();

        // Add the same highlight effect as regular balloons
        const gradient = ctx.createRadialGradient(
            x - balloon.radius / 3,
            y - balloon.radius / 3,
            balloon.radius / 10,
            x,
            y,
            balloon.radius
        );
        gradient.addColorStop(0, "rgba(255, 255, 255, 0.4)");
        gradient.addColorStop(1, "rgba(255, 255, 255, 0)");
        ctx.fillStyle = gradient;
        ctx.fill();

        // Add subtle negative indicators that require careful observation

        // 1. Small minus sign (-) in the center
        ctx.beginPath();
        ctx.moveTo(x - 6, y);
        ctx.lineTo(x + 6, y);
        ctx.strokeStyle = "rgba(255, 0, 0, 0.7)";
        ctx.lineWidth = 2;
        ctx.stroke();

        // 2. Very thin red ring around the balloon (barely visible)
        ctx.beginPath();
        ctx.arc(x, y, balloon.radius * 0.98, 0, Math.PI * 2);
        ctx.strokeStyle = "rgba(255, 0, 0, 0.3)";
        ctx.lineWidth = 1;
        ctx.stroke();

        // 3. Tiny red dots at cardinal points (N, S, E, W)
        const dotRadius = 1.5;
        const dotDistance = balloon.radius * 0.75;

        // North dot
        ctx.beginPath();
        ctx.arc(x, y - dotDistance, dotRadius, 0, Math.PI * 2);
        ctx.fillStyle = "rgba(255, 0, 0, 0.7)";
        ctx.fill();

        // South dot
        ctx.beginPath();
        ctx.arc(x, y + dotDistance, dotRadius, 0, Math.PI * 2);
        ctx.fillStyle = "rgba(255, 0, 0, 0.7)";
        ctx.fill();

        // East dot
        ctx.beginPath();
        ctx.arc(x + dotDistance, y, dotRadius, 0, Math.PI * 2);
        ctx.fillStyle = "rgba(255, 0, 0, 0.7)";
        ctx.fill();

        // West dot
        ctx.beginPath();
        ctx.arc(x - dotDistance, y, dotRadius, 0, Math.PI * 2);
        ctx.fillStyle = "rgba(255, 0, 0, 0.7)";
        ctx.fill();
    }

    drawPopEffects() {
        const ctx = this.ctx;

        this.popEffects.forEach((effect) => {
            effect.particles.forEach((particle) => {
                if (particle.life > 0) {
                    ctx.beginPath();
                    ctx.arc(particle.x, particle.y, 3, 0, Math.PI * 2);
                    ctx.fillStyle = particle.color;
                    ctx.globalAlpha = particle.life;
                    ctx.fill();
                    ctx.globalAlpha = 1;
                }
            });
        });
    }

    drawGameInfo() {
        const ctx = this.ctx;
        const timePlayed = Math.floor((Date.now() - this.gameStartTime) / 1000);

        ctx.font = "16px Arial";
        ctx.fillStyle = "black";
        ctx.textAlign = "left";
        ctx.fillText(`Level: ${this.level}`, 10, 30);
        ctx.fillText(`Speed: ${this.speedMultiplier.toFixed(1)}x`, 10, 50);
        ctx.fillText(`Time: ${timePlayed}s`, 10, 70);
    }

    draw() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        this.balloons.forEach((balloon) => this.drawBalloon(balloon));
        this.drawPopEffects();
        this.drawGameInfo();
    }

    updateScore(eventData) {
        console.log("Updating score:", eventData);



        if (eventData.negative_hit) {
            this.negativeHits++;
            // Handle scoring update
            this.score += eventData.score;
        } else {
            if (eventData.favorite_color_bonus) {
                this.bonusHits++;
            } else {
                // Handle scoring update
                this.score += eventData.score;
                this.regularHits++;
            }
            // Add animations for positive scores
            const scoreElement = document.getElementById("score");
            if (scoreElement) {
                if (eventData.favorite_color_bonus) {
                    scoreElement.classList.add("bonus");
                    setTimeout(() => scoreElement.classList.remove("bonus"), 1000);
                }
            }
        }

        // Update UI
        const scoreElement = document.getElementById("score");
        const bonusElement = document.getElementById("bonusHits");
        const regularElement = document.getElementById("regularHits");
        const negativeElement = document.getElementById("negativeHits");

        if (scoreElement) scoreElement.textContent = `Score: ${this.score}`;
        if (bonusElement) bonusElement.textContent = `Bonus Hits: ${this.bonusHits}`;
        if (regularElement) regularElement.textContent = `Regular Hits: ${this.regularHits}`;
        if (negativeElement) negativeElement.textContent = `Negative Hits: ${this.negativeHits}`;
    }

    stopGame() {
        this.isActive = false;
        this.balloons = [];
        this.popEffects = [];
    }

    gameLoop() {
        if (this.isActive) {
            const currentTime = Date.now();
            if (currentTime - this.lastSpawnTime > this.spawnInterval) {
                this.createBalloon();
                this.lastSpawnTime = currentTime;
            }

            this.updateBalloons();
        }

        this.draw();
        requestAnimationFrame(() => this.gameLoop());
    }
}