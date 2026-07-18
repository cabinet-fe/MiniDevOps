<template>
  <div class="brand-logo" aria-label="Bedrock 磐石">
    <svg
      class="brand-logo__mark"
      viewBox="0 0 40 40"
      width="32"
      height="32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <!-- Stratified load-bearing stone: verdigris slabs + brass seam -->
      <path
        d="M6 28 L20 34 L34 28 L34 22 L20 28 L6 22 Z"
        stroke="currentColor"
        stroke-width="1.6"
        stroke-linejoin="round"
        class="brand-logo__verdigris"
      />
      <path
        d="M6 20 L20 26 L34 20 L34 14 L20 20 L6 14 Z"
        stroke="currentColor"
        stroke-width="1.6"
        stroke-linejoin="round"
        class="brand-logo__verdigris"
      />
      <path
        d="M6 12 L20 18 L34 12 L28 8 L20 11 L12 8 Z"
        stroke="currentColor"
        stroke-width="1.6"
        stroke-linejoin="round"
        class="brand-logo__verdigris"
      />
      <path
        d="M12 8 L20 11 L28 8"
        stroke="currentColor"
        stroke-width="1.2"
        stroke-linecap="round"
        class="brand-logo__brass"
      />
      <path
        d="M20 11 V34"
        stroke="currentColor"
        stroke-width="1"
        stroke-linecap="round"
        opacity="0.55"
        class="brand-logo__brass"
      />
    </svg>
    <div class="brand-logo__text">
      <span class="brand-logo__name">BEDROCK</span>
      <span class="brand-logo__sub">磐石</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.brand-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  color: fn.use-var(color, primary);
}

.brand-logo__mark {
  flex-shrink: 0;
  display: block;
}

.brand-logo__verdigris {
  color: fn.use-var(color, primary);
}

.brand-logo__brass {
  color: #b08a4a;
}

.brand-logo__text {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
  line-height: 1.15;
}

.brand-logo__name {
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.14em;
  color: fn.use-var(text-color, title);
}

.brand-logo__sub {
  font-size: 11px;
  letter-spacing: 0.28em;
  color: fn.use-var(text-color, second);
}
</style>
