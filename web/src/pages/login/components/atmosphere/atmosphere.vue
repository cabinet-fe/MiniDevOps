<template>
  <div class="atmosphere" aria-hidden="true">
    <div class="paper-base" />

    <!-- 远山水墨，三层淡墨由近及远 -->
    <svg class="mountains" viewBox="0 0 1440 900" preserveAspectRatio="xMidYMax slice">
      <defs>
        <linearGradient id="ink-far" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#9aa392" stop-opacity="0.3" />
          <stop offset="100%" stop-color="#9aa392" stop-opacity="0" />
        </linearGradient>
        <linearGradient id="ink-mid" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#77836f" stop-opacity="0.36" />
          <stop offset="100%" stop-color="#77836f" stop-opacity="0" />
        </linearGradient>
        <linearGradient id="ink-near" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#55624f" stop-opacity="0.42" />
          <stop offset="100%" stop-color="#55624f" stop-opacity="0" />
        </linearGradient>
        <filter id="ink-soft" x="-10%" y="-10%" width="120%" height="120%">
          <feGaussianBlur stdDeviation="3" />
        </filter>
        <filter id="cloud-soft" x="-40%" y="-40%" width="180%" height="180%">
          <feGaussianBlur stdDeviation="14" />
        </filter>
      </defs>

      <g filter="url(#ink-soft)">
        <path
          fill="url(#ink-far)"
          d="M0 660 C 140 590 240 560 380 600 C 520 640 620 520 780 570 C 920 615 1040 545 1200 585 C 1310 610 1390 575 1440 595 L1440 900 L0 900 Z"
        />
        <path
          fill="url(#ink-mid)"
          d="M0 740 C 130 680 260 655 420 690 C 580 725 700 640 880 680 C 1030 712 1180 660 1320 690 C 1390 705 1420 695 1440 690 L1440 900 L0 900 Z"
        />
        <path
          fill="url(#ink-near)"
          d="M0 820 C 170 765 330 745 520 775 C 710 805 860 745 1060 775 C 1220 798 1340 770 1440 780 L1440 900 L0 900 Z"
        />
      </g>

      <!-- 云雾 -->
      <g fill="#faf7ee" opacity="0.7" filter="url(#cloud-soft)">
        <ellipse cx="320" cy="640" rx="190" ry="22" />
        <ellipse cx="900" cy="600" rx="230" ry="26" />
        <ellipse cx="1260" cy="680" rx="170" ry="20" />
      </g>
    </svg>

    <div class="carving">磐</div>
    <div class="mist" />
    <div class="vignette" />
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

/* 宣纸底：暖白渐变 + 细腻纤维纹 */
.paper-base {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 90% 70% at 50% 30%, #faf7ee 0%, transparent 70%),
    repeating-linear-gradient(95deg, transparent 0 6px, rgb(64 54 32 / 1.2%) 6px 7px),
    repeating-linear-gradient(4deg, transparent 0 9px, rgb(64 54 32 / 1%) 9px 10px),
    linear-gradient(180deg, #f6f2e6 0%, #f1ede0 55%, #eae4d2 100%);
}

.mountains {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

/* 淡墨大字，如碑刻隐约 */
.carving {
  position: absolute;
  left: 50%;
  top: 42%;
  translate: -50% -50%;
  font-family: "Songti SC", "STSong", "SimSun", "Noto Serif CJK SC", serif;
  font-size: min(58vw, 480px);
  font-weight: 700;
  line-height: 1;
  color: rgb(61 107 88 / 5%);
  letter-spacing: 0.08em;
  user-select: none;
  animation: carving-breathe 14s ease-in-out infinite;
}

/* 流岚轻雾 */
.mist {
  position: absolute;
  inset: -8%;
  background:
    radial-gradient(ellipse 45% 30% at 28% 66%, rgb(255 255 255 / 55%), transparent 70%),
    radial-gradient(ellipse 40% 28% at 72% 30%, rgb(255 255 255 / 40%), transparent 65%);
  animation: mist-drift 20s ease-in-out infinite alternate;
}

/* 四周淡淡晕影，收拢视线 */
.vignette {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 70% at 50% 42%, transparent 55%, rgb(120 105 75 / 14%) 100%),
    linear-gradient(180deg, rgb(120 105 75 / 8%) 0%, transparent 14%);
}

@keyframes mist-drift {
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
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 480px) {
  .carving {
    font-size: 70vw;
  }
}
</style>
