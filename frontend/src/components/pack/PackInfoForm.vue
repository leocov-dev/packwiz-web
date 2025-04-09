<script lang="ts">

import type {LoaderVersions} from "@/stores/cache.ts";

export interface PackInfoFormData {
  slug: string,
  name?: string,
  packVersion?: string,
  description?: string,
  minecraftVersion?: string,
  loader: {
    name?: keyof LoaderVersions,
    version?: string,
  },
  acceptableVersions: string[],
}
</script>

<script setup lang="ts">
import Loader from "@/components/forms/Loader.vue";
import MinecraftVersion from "@/components/forms/MinecraftVersion.vue";

const data = defineModel<PackInfoFormData>('data', {required: true})
const loading = defineModel<boolean>('loading', {required: true})
const emit = defineEmits(["submit-data", "cancel-op"])

const {title, acceptText} = defineProps({
  title: {type: String, required: true},
  acceptText: {type: String, required: true},
})

const isValid = ref(false)
</script>

<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <h1 class="me-5">
        {{ title }}
      </h1>
    </v-card-title>

    <v-form
      v-model="isValid"
      validate-on="eager"
      class="ma-6"
      @submit.prevent="emit('submit-data')"
    >
      <v-row>
        <v-col class="v-col-md-9 v-col-sm-8">
          <SlugAndName
            v-model:slug="data.slug"
            v-model:name="data.name"
          />
        </v-col>
        <v-col class="v-col-md-3 v-col-sm-4">
          <v-text-field
            v-model="data.packVersion"
            label="Pack Version"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-divider />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <MinecraftVersion
            v-model:version="data.minecraftVersion"
            :include-latest="true"
          />
        </v-col>
        <v-col>
          <AcceptableVersions
            v-model:acceptable-versions="data.acceptableVersions"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-divider />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <Loader
            v-model:loader="data.loader.name"
            v-model:version="data.loader.version"
            :minecraft-version="data.minecraftVersion"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-divider />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-textarea
            v-model="data.description"
            label="Description"
          />
        </v-col>
      </v-row>

      <div class="d-flex justify-end">
        <v-btn
          :disabled="loading"
          text="Cancel"
          class="me-6"
          @click="emit('cancel-op')"
        />
        <v-btn
          :disabled="loading || !isValid"
          :text="acceptText"
          type="submit"
          color="primary"
        />
      </div>
    </v-form>

    <v-overlay
      v-model="loading"
      class="align-center justify-center"
      persistent
      contained
    >
      <v-progress-circular
        color="primary"
        size="64"
        indeterminate
      />
    </v-overlay>
  </v-card>
</template>
