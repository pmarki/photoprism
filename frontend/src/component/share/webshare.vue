<template>
  <v-dialog
    :model-value="showConfirmationDialog"
    lazy
    persistent
    max-width="350"
    class="p-confirm-dialog"
    @keydown.esc="cancel"
  >
    <v-card elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-layout row wrap>
          <div class="xs3 text-center">
            <v-icon size="54" color="secondary-dark lighten-1" icon="mdi-share-variant"></v-icon>
          </div>
          <div class="xs9 text-left align-self-center">
            <div class="subheading pr-1">
              <translate>Download complete. Ready for sharing?</translate>
            </div>
          </div>
          <div class="xs12 text-right pt-3">
            <v-btn
              color="secondary-light"
              class="action-cancel compact"
              @click.stop="cancel"
            >
              <translate key="Cancel">Cancel</translate>
            </v-btn>
            <v-btn
              color="primary-button"
              class="action-confirm compact"
              @click.stop="webShareDialogInitiated"
            >
              <translate key="Share">Share</translate>
            </v-btn>
          </div>
        </v-layout>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
import Photo from "model/photo";
import Util from "common/util";
import Notify from "common/notify";
import Api from "common/api";
export default {
  name: "PWebshareDialog",
  props: {
    show: {
      type: Boolean,
      default: false,
    },
    items: {
      type: Object,
      required: true,
    },
    icon: {
      type: String,
      default: "share",
    },
  },
  watch: {
    show: function (val) {
      if (val) {
        this.onShow();
      }
    },
  },
  data() {
    return {
      showConfirmationDialog: false,
      webshareData: [],
    };
  },
  methods: {
    cancel() {
      this.showConfirmationDialog = false;
      this.$emit("cancel");
    },
    onShow() {
      // Resolve selection into Photo objects and download them as blobs
      const photos = this.items.photos.map((uid) =>
        new Photo().find(uid).then((p) =>
          Api.get(p.getWebshareDownloadUrl(), { responseType: "blob" }).then(
            (resp) => {
              p.Blob = resp.data;
              return p;
            }
          )
        )
      );
      // Wait for all downloads, then open native browser share dialog
      Promise.all(photos)
        .then((blobs) => {
          const filesArray = blobs.map((p) =>
            Util.JSFileForWebshare(p.Blob, p.getWebshareFile())
          );
          const webshareData = {
            files: filesArray,
          };
          this.webshareData = webshareData;
          return navigator.share(webshareData);
        })
        .catch((e) => {
          if (e.name === "AbortError") {
            // Cancelled by user
            this.$emit("cancel");
          } else if (e.name === "NotAllowedError") {
            // Sharing requires a transient activation and might fail with a NotAllowedError.
            // Show dialog to create transient activation and try again.
            this.showConfirmationDialog = true;
          } else {
            this.$emit("failed");
            this.$notify.error(e.message);
          }
        })
        .finally(() => {
          if (!this.showConfirmationDialog) {
            this.webshareData = {};
          }
          this.$emit("completed");
        });
      Notify.success(this.$gettext("Sharing…"));
    },
    webShareDialogInitiated() {
      this.showConfirmationDialog = false;
      navigator
        .share(this.webshareData)
        .catch((e) => {
          if (e.name === "AbortError") {
            this.$emit("cancel");
          } else {
            this.$emit("failed");
            this.$notify.error(e.message);
          }
        })
        .finally(() => {
          this.webshareData = {};
        });
    },
  },
};
</script>