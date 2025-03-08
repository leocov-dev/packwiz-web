<script setup lang="ts">

const slug = defineModel<string>('slug', {required: true})
const name = defineModel<string>('name', {required: true})

const {slugLocked} = defineProps<{ slugLocked?: boolean }>()

const rules = {
  slugRequired: (value: string) => !!value || "Slug is required",
  nameRequired: (value: string) => !!value || "Name is required",
}

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
  if (!slugLocked) {
    slug.value = slugFromName(name.value)
  }
})

</script>

<template>
  <div class="d-flex">
    <v-text-field
      v-model="slug"
      :rules="[rules.slugRequired]"
      label="Slug"
      disabled
      class="me-6"
    />
    <v-text-field
      v-model="name"
      :rules="[rules.nameRequired]"
      label="Name"
    />
  </div>
</template>

