<script setup lang="ts">

const slug = defineModel<string>('slug', {required: true})
const name = defineModel<string>('name', {required: true})


const slugFromName = (name: string) => {
  const spaceSeparated = name.replace(/[^a-zA-Z0-9-_\.]/g, ' ')

  return spaceSeparated
    .split(' ')
    .filter(word => word.length > 0) // Remove empty strings
    .map((word, index) => {
      if (index === 0) {
        return word.toLowerCase()
      }
      return word.charAt(0).toUpperCase() + word.slice(1).toLowerCase()
    })
    .join('')

}

watch(name, () => {
  slug.value = slugFromName(name.value)
})

</script>

<template>
  <div class="d-flex">
    <v-text-field
      v-model="slug"
      label="Slug"
      disabled
      class="me-6"
    />
    <v-text-field
      v-model="name"
      label="Name"
    />
  </div>
</template>

