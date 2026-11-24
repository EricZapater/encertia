<script setup lang="ts">
import { computed } from 'vue'
import { formatQuestionText } from '@/utils/textFormat'

const props = withDefaults(
  defineProps<{
    text: string | null | undefined
    imageUrl?: string | null
    customClass?: string
  }>(),
  {
    text: '',
    imageUrl: null,
    customClass: ''
  }
)

const renderedHtml = computed(() => {
  return formatQuestionText(props.text, props.imageUrl)
})
</script>

<template>
  <div
    class="formatted-question-root"
    :class="customClass"
    v-html="renderedHtml"
    data-testid="formatted-question-text"
  />
</template>

<style>
.formatted-question-root {
  line-height: 1.4;
  word-break: break-word;
}

.formatted-question-root .formatted-h1 {
  font-size: 1.35em;
  font-weight: 900;
  margin: 0.25em 0;
  line-height: 1.25;
}

.formatted-question-root .formatted-h2 {
  font-size: 1.2em;
  font-weight: 800;
  margin: 0.25em 0;
  line-height: 1.25;
}

.formatted-question-root .formatted-h3 {
  font-size: 1.1em;
  font-weight: 700;
  margin: 0.2em 0;
  line-height: 1.25;
}

.formatted-question-root strong {
  font-weight: 900;
}

.formatted-question-root em {
  font-style: italic;
}

.formatted-question-root u {
  text-decoration: underline;
  text-underline-offset: 4px;
}

.formatted-question-root .inline-question-img-wrapper {
  display: block;
  margin: 0.85rem auto;
  text-align: center;
  max-width: 100%;
}

.formatted-question-root .inline-question-img {
  max-height: 40vh;
  max-width: 100%;
  border-radius: 0.75rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
  object-fit: contain;
  vertical-align: middle;
}
</style>
