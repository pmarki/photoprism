<template>
  <v-dialog
    :model-value="visible"
    persistent
    scrim
    max-width="350"
    class="p-dialog p-album-create-dialog"
    @keydown.esc.exact="close"
  >
    <v-form ref="form" class="p-photo-upload" validate-on="invalid-input" tabindex="1" @submit.prevent="onSubmit">
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon icon="mdi-book-plus" size="54" color="primary"></v-icon>
          <p class="text-subtitle-1">{{ $gettext(title()) }}</p>
        </v-card-title>
        <v-card-text>
          <v-text-field label="Name" v-model="name"></v-text-field>
        </v-card-text>
        <v-card-actions class="action-buttons mt-1">
          <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn color="highlight" variant="flat" class="action-confirm" :disabled="!name" @click.stop="confirm">
            {{ $gettext(`Save`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
      </v-form>
  </v-dialog>
</template>
<script>
import $api from "common/api";
import $notify from "common/notify";
import Folder from "../../../model/folder";
import Album from "../../../model/album";

export default {
  name: "PAlbumCreateDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["close", "confirm"],
  data() {
    return {
      name: null,
    };
  },
  computed: {
    isFolder() {
      return this.$route.name.startsWith("folder");
    }
  },
  methods: {
    title() {
      if (this.isFolder) {
        return "Create new folder";
      } else {
        return "Create new album";
      }
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      if (this.isFolder) {
        // TODO add path
        Folder.createNew(this.name);
        this.$notify.success(this.$gettext("Folder created"));
        this.$emit("close");

        return;
      }

      let title = this.name;
      // TODO check if not exist

      const album = new Album({ Title: title, Favorite: false });

      album.save().then(() => this.$notify.success(this.$gettext("Album created")));
      this.$emit("close");
    },
    onSubmit() {

    }
  },
};
</script>
