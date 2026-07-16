<template>
  <div class="atmosphere" aria-hidden="true">
    <div class="abyss" />
    <div class="strata" />
    <div class="veins" />
    <svg class="masonry" viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice">
      <defs>
        <linearGradient id="block-face" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#2f3d35" />
          <stop offset="40%" stop-color="#1c2620" />
          <stop offset="100%" stop-color="#0a0f0c" />
        </linearGradient>
        <linearGradient id="block-hi" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#4a5c52" stop-opacity="0.35" />
          <stop offset="100%" stop-color="#000" stop-opacity="0" />
        </linearGradient>
        <linearGradient id="crack-glow" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#5a8a72" stop-opacity="0" />
          <stop offset="40%" stop-color="#5a8a72" stop-opacity="0.55" />
          <stop offset="100%" stop-color="#a67c3d" stop-opacity="0" />
        </linearGradient>
        <filter id="soft-edge" x="-20%" y="-20%" width="140%" height="140%">
          <feGaussianBlur stdDeviation="0.6" />
        </filter>
        <pattern id="grit" width="120" height="120" patternUnits="userSpaceOnUse">
          <circle cx="8" cy="14" r="0.7" fill="#9bb5a6" opacity="0.18" />
          <circle cx="42" cy="66" r="0.5" fill="#c4a36a" opacity="0.14" />
          <circle cx="88" cy="28" r="0.6" fill="#9bb5a6" opacity="0.12" />
          <circle cx="110" cy="96" r="0.5" fill="#8a5a42" opacity="0.1" />
          <circle cx="60" cy="48" r="0.4" fill="#d8d2c4" opacity="0.08" />
        </pattern>
        <path id="blk-a" d="M-20 80 L280 40 L520 110 L480 260 L240 290 L-20 220 Z" />
        <path id="blk-b" d="M500 70 L820 20 L1120 90 L1080 250 L760 290 L520 220 Z" />
        <path id="blk-c" d="M1100 60 L1480 100 L1480 280 L1180 300 L1060 220 Z" />
        <path id="blk-d" d="M-20 300 L260 270 L500 340 L460 520 L200 560 L-20 480 Z" />
        <path id="blk-e" d="M480 310 L860 250 L1180 330 L1140 520 L780 560 L500 480 Z" />
        <path id="blk-f" d="M1160 300 L1480 340 L1480 560 L1220 580 L1120 470 Z" />
        <path id="blk-g" d="M-20 560 L300 520 L560 610 L520 820 L180 860 L-20 760 Z" />
        <path id="blk-h" d="M540 580 L900 520 L1240 600 L1200 820 L820 870 L560 780 Z" />
        <path id="blk-i" d="M1220 560 L1480 600 L1480 900 L1180 900 L1120 740 Z" />
      </defs>

      <g fill="url(#block-face)">
        <use href="#blk-a" />
        <use href="#blk-b" />
        <use href="#blk-c" />
        <use href="#blk-d" />
        <use href="#blk-e" />
        <use href="#blk-f" />
        <use href="#blk-g" />
        <use href="#blk-h" />
        <use href="#blk-i" />
      </g>
      <g fill="url(#block-hi)">
        <use href="#blk-a" />
        <use href="#blk-b" />
        <use href="#blk-d" />
        <use href="#blk-g" />
      </g>
      <g fill="none" stroke="#8aa396" stroke-opacity="0.16" stroke-width="1.2">
        <use href="#blk-a" />
        <use href="#blk-b" />
        <use href="#blk-d" />
        <use href="#blk-e" />
        <use href="#blk-g" />
        <use href="#blk-h" />
      </g>

      <g fill="none" stroke="url(#crack-glow)" stroke-width="2" filter="url(#soft-edge)">
        <path d="M260 270 L500 340 L860 250 L1180 330" />
        <path d="M200 560 L500 480 L780 560 L1140 520 L1220 580" />
        <path d="M280 40 L280 270 M820 20 L860 250 M1120 90 L1180 330" />
        <path d="M300 520 L300 820 M900 520 L900 820" />
        <path d="M40 200 L200 180 L240 290 L90 340 Z" opacity="0.5" />
      </g>

      <g fill="none" stroke="#6b9a82" stroke-opacity="0.14" stroke-width="1">
        <polygon points="720,120 860,240 780,360 660,360 580,240" />
        <polygon points="720,180 800,250 720,320 640,250" />
        <circle cx="720" cy="250" r="18" />
        <path d="M200 700 L280 640 L360 720 L260 780 Z" />
        <path d="M1080 180 L1180 140 L1240 220 L1140 260 Z" />
      </g>

      <g fill="none" stroke="#6a5634" stroke-opacity="0.28" stroke-width="6">
        <path d="M90 0 V900M1350 0 V900" />
      </g>
      <g fill="none" stroke="#4a7a64" stroke-opacity="0.2" stroke-width="2">
        <path d="M90 0 V900M1350 0 V900M0 160 H1440M0 740 H1440" />
      </g>

      <g fill="none" stroke="#b08a4a" stroke-opacity="0.35" stroke-width="1.5">
        <path d="M40 120 H160 V40M1280 40 V120 H1400M40 780 H160 V860M1280 860 V780 H1400" />
      </g>

      <rect width="1440" height="900" fill="url(#grit)" />
    </svg>

    <div class="carving">磐</div>
    <div class="haze" />
    <div class="vignette" />
    <div class="scan" />
  </div>
</template>

<style scoped lang="scss">
.atmosphere {
  position: absolute;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
}

.abyss {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 60% at 50% 40%, #152019 0%, transparent 70%),
    linear-gradient(180deg, #0e1411 0%, var(--ink) 48%, #040605 100%);
}

.strata {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(115deg, rgb(74 122 100 / 16%) 0%, transparent 40%),
    linear-gradient(300deg, var(--rust-mist) 0%, transparent 45%),
    repeating-linear-gradient(
      8deg,
      transparent 0 70px,
      rgb(255 255 255 / 2.2%) 70px 71px,
      transparent 71px 148px,
      rgb(0 0 0 / 22%) 148px 151px
    ),
    repeating-linear-gradient(98deg, transparent 0 90px, rgb(0 0 0 / 14%) 90px 92px);
}

.veins {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 30% 20% at 22% 68%, rgb(74 122 100 / 28%), transparent 70%),
    radial-gradient(ellipse 25% 18% at 78% 30%, rgb(176 138 74 / 14%), transparent 70%),
    radial-gradient(ellipse 40% 30% at 50% 90%, rgb(90 50 36 / 20%), transparent 70%);
  animation: haze-drift 22s ease-in-out infinite alternate;
}

.masonry {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0.95;
}

.carving {
  position: absolute;
  left: 50%;
  top: 46%;
  translate: -50% -50%;
  font-family: "Songti SC", "STSong", "SimSun", "Noto Serif CJK SC", serif;
  font-size: min(62vw, 520px);
  font-weight: 700;
  line-height: 1;
  color: rgb(212 207 195 / 3.5%);
  letter-spacing: 0.08em;
  text-shadow:
    0 2px 0 rgb(0 0 0 / 35%),
    0 0 80px rgb(74 122 100 / 8%);
  user-select: none;
  animation: carving-breathe 14s ease-in-out infinite;
}

.haze {
  position: absolute;
  inset: -8%;
  background:
    radial-gradient(ellipse 50% 35% at 30% 70%, rgb(20 40 32 / 55%), transparent 70%),
    radial-gradient(ellipse 45% 30% at 70% 25%, rgb(40 32 18 / 30%), transparent 65%);
  animation: haze-drift 18s ease-in-out infinite alternate-reverse;
  mix-blend-mode: multiply;
}

.vignette {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 75% 65% at 50% 42%, transparent 42%, rgb(0 0 0 / 42%) 100%),
    linear-gradient(
      180deg,
      rgb(0 0 0 / 18%) 0%,
      transparent 16%,
      transparent 78%,
      rgb(0 0 0 / 38%) 100%
    );
}

.scan {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(180deg, transparent 0 3px, rgb(0 0 0 / 4%) 3px 4px);
  opacity: 0.45;
}

@keyframes haze-drift {
  from {
    transform: translate3d(-1.2%, 0.4%, 0) scale(1.02);
  }

  to {
    transform: translate3d(1.4%, -0.8%, 0) scale(1.05);
  }
}

@keyframes carving-breathe {
  0%,
  100% {
    opacity: 0.85;
    filter: brightness(0.95);
  }

  50% {
    opacity: 1;
    filter: brightness(1.15);
  }
}

@media (max-width: 480px) {
  .carving {
    font-size: 70vw;
  }
}
</style>
