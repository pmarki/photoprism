<template>
  <div id="p-navigation" class="p-page__navigation" :class="{ 'sidenav-visible': drawer }">
    <template v-if="visible">
      <template v-if="$vuetify.display.xlAndDown || $vuetify.display.xlAndUp">
        <v-toolbar
          position="fixed"
          flat
          density="compact"
          color="navigation"
          class="nav-small elevation-2"
          @click.stop.prevent
        >
          <v-btn icon variant="text" class="bg-transparent nav-logo" @click.stop.prevent="toggleDrawer">
            <img :src="appIcon" :alt="appName" :class="{ 'animate-hue': indexing }" />
          </v-btn>
          <v-toolbar-title class="nav-toolbar-title">
            <slot name="title">
              <span :class="{ clickable: auth }" @click.stop.prevent="toggleDrawer">{{ page.title }}</span>
            </slot>
          </v-toolbar-title>
          <slot name="menu-icons"></slot>
          <slot name="menu-actions"></slot>
        </v-toolbar>
      </template>
      <template v-else-if="!auth">
        <v-toolbar
          flat
          :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
          color="navigation"
          class="nav-small"
        >
          <v-btn icon variant="text" class="bg-transparent nav-logo">
            <img :src="appIcon" :alt="appName" />
          </v-btn>
          <v-toolbar-title class="nav-toolbar-title">
            {{ page.title }}
          </v-toolbar-title>

          <v-btn
            icon="mdi-dots-vertical"
            variant="text"
            class="nav-mobile-menu-trigger elevation-0"
            :ripple="false"
            @click.stop.prevent="speedDial = true"
          ></v-btn>
        </v-toolbar>
      </template>
      <v-navigation-drawer
        v-if="auth"
        v-model="drawer"
        :rail="isMini"
        temporary
        color="navigation"
        class="nav-sidebar navigation"
      >
        <div class="nav-container">
          <v-toolbar flat :density="$vuetify.display.smAndDown ? 'compact' : 'default'">
            <v-list
              class="navigation-home elevation-0"
              bg-color="navigation-home"
              width="100%"
              density="compact"
              tabindex="-1"
              @click.capture="toggleDrawer"
            >
              <v-list-item class="px-3" :elevation="0" :ripple="false" @click.stop.prevent="goHome">
                <template #prepend>
                  <div class="v-avatar bg-transparent nav-logo">
                    <a :href="siteUrl" tabindex="-1" @click.stop.prevent="goHome">
                      <img :src="appIcon" :alt="appName" :class="{ 'animate-hue': indexing }" />
                    </a>
                  </div>
                </template>
                <template #append>
                  <v-btn
                    icon
                    variant="text"
                    :elevation="0"
                    tabindex="-1"
                    class="nav-minimize hidden-sm-and-down"
                    :ripple="false"
                    :title="$gettext('Minimize')"
                    @click.stop.prevent="toggleIsMini()"
                  >
                    <v-icon :icon="rtl ? 'mdi-chevron-right' : 'mdi-chevron-left'"></v-icon>
                  </v-btn>
                </template>
                <v-list-item-title class="nav-toolbar-title">
                  <a :href="siteUrl">{{ appName }}</a>
                </v-list-item-title>
              </v-list-item>
            </v-list>
          </v-toolbar>

          <v-list
            nav
            class="nav-menu"
            bg-color="navigation"
            color="primary"
            open-strategy="single"
            :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
            tabindex="-1"
          >
            <v-list-item v-if="isMini && !isRestricted" class="nav-expand" @click.stop="toggleIsMini()">
              <v-icon :icon="rtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" class="ma-auto"></v-icon>
            </v-list-item>

            <v-list-item
              v-if="isMini && $config.feature('search')"
              to="/browse"
              variant="text"
              class="nav-browse"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-magnify</v-icon>
            </v-list-item>
            <div v-else-if="!isMini && $config.feature('search')">
              <v-list-item to="/browse" variant="text" class="nav-browse activator" @click.stop="">
                <v-list-item-title class="nav-menu-item">
                  <p class="nav-item-title">
                    {{ $gettext(`Show all`) }}
                  </p>
                  <span v-if="config.count.all > 0" class="nav-count-group">{{ config.count.all }}</span>
                </v-list-item-title>
              </v-list-item>

              <v-list-group>
                <template #activator="{ props }">
                  <v-list-item v-bind="props" variant="text" class="nav-browse activator-parent" @click.stop="">
                    <v-icon>mdi-magnify</v-icon>
                  </v-list-item>
                </template>
                <v-list-item :to="{ name: 'videos' }" variant="text" class="nav-video nav-videos" @click.stop="">
                  <v-list-item-title :class="`nav-menu-item menu-item`">
                    {{ $gettext(`Videos`) }}
                  </v-list-item-title>
                  <span v-show="config.count.videos > 0" class="nav-count-item">{{ config.count.videos }}</span>
                </v-list-item>
                <v-list-item
                  v-if="canManagePhotos"
                  v-show="$config.feature('review')"
                  to="/review"
                  variant="text"
                  class="nav-review"
                  @click.stop=""
                >
                  <v-list-item-title :class="`nav-menu-item menu-item`">
                    {{ $gettext(`Review`) }}
                  </v-list-item-title>
                  <span v-show="config.count.review > 0" class="nav-count-item">{{ config.count.review }}</span>
                </v-list-item>

                <v-list-item
                  v-if="canAccessPrivate"
                  v-show="$config.feature('private')"
                  to="/private"
                  variant="text"
                  class="nav-private"
                  @click.stop=""
                >
                  <v-list-item-title :class="`nav-menu-item menu-item`">
                    {{ $gettext(`Private`) }}
                  </v-list-item-title>
                  <span v-show="config.count.private > 0" class="nav-count-item">{{ config.count.private }}</span>
                </v-list-item>

                <v-list-item
                  v-show="$config.feature('archive')"
                  to="/archive"
                  variant="text"
                  class="nav-archive"
                  @click.stop=""
                >
                  <v-list-item-title :class="`nav-menu-item menu-item`">
                    {{ $gettext(`Archive`) }}
                  </v-list-item-title>
                  <span v-show="config.count.archived > 0" class="nav-count-item">{{ config.count.archived }}</span>
                </v-list-item>
              </v-list-group>
            </div>

            <v-list-item
              v-if="isMini"
              v-show="$config.feature('albums')"
              to="/albums"
              variant="text"
              class="nav-albums"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-bookmark</v-icon>
            </v-list-item>
            <div v-else-if="!isMini" v-show="$config.feature('albums')">
              <v-list-item to="/albums" variant="text" class="nav-albums activator" @click.stop="">
                <v-list-item-title class="nav-menu-item">
                  <p class="nav-item-title">
                    {{ $gettext(`Albums`) }}
                  </p>
                  <span v-if="config.count.albums > 0" class="nav-count-group">{{ config.count.albums }}</span>
                </v-list-item-title>
              </v-list-item>

              <v-list-group>
                <template #activator="{ props }">
                  <v-list-item
                    v-bind="props"
                    variant="text"
                    class="nav-albums activator-parent"
                    :ripple="false"
                    @click.stop=""
                  >
                    <v-icon>mdi-bookmark</v-icon>
                  </v-list-item>
                </template>

                <v-list-item to="/unsorted" variant="text" class="nav-unsorted">
                  <v-list-item-title :class="`nav-menu-item menu-item`">
                    {{ $gettext(`Unsorted`) }}
                  </v-list-item-title>
                </v-list-item>
              </v-list-group>
            </div>

            <v-list-item
              v-if="isMini && $config.feature('videos')"
              to="/media"
              variant="text"
              class="nav-media"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-play-circle</v-icon>
            </v-list-item>

            <v-list-item
              v-if="isMini && $config.feature('people') && (canManagePeople || config.count.people > 0)"
              :to="{ name: 'people' }"
              variant="text"
              class="nav-people"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-account</v-icon>
            </v-list-item>
            <v-list-item
              v-else-if="!isMini && $config.feature('people') && (canManagePeople || config.count.people > 0)"
              :to="{ name: 'people' }"
              variant="text"
              class="nav-people"
              :ripple="false"
              @click.stop=""
            >
              <v-list-item-title class="nav-menu-item">
                <v-icon>mdi-account</v-icon>
                <p class="nav-item-title">
                  {{ $gettext(`People`) }}
                </p>
              </v-list-item-title>
              <span v-show="config.count.people > 0" class="nav-count-item">{{ config.count.people }}</span>
            </v-list-item>

            <v-list-item
              v-if="isMini && $config.feature('favorites')"
              :to="{ name: 'favorites' }"
              variant="text"
              class="nav-favorites"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-star</v-icon>
            </v-list-item>
            <v-list-item
              v-else-if="!isMini && $config.feature('favorites')"
              :to="{ name: 'favorites' }"
              variant="text"
              class="nav-favorites"
              :ripple="false"
              @click.stop=""
            >
              <v-list-item-title class="nav-menu-item">
                <v-icon>mdi-star</v-icon>
                <p class="nav-item-title">
                  {{ $gettext(`Favorites`) }}
                </p>
              </v-list-item-title>
              <span v-show="config.count.favorites > 0" class="nav-count-item">{{ config.count.favorites }}</span>
            </v-list-item>

            <template v-if="isRestricted && $config.feature('places')">
              <v-list-item
                v-if="isMini"
                :to="{ name: 'states' }"
                variant="text"
                class="nav-states nav-regions"
                :ripple="false"
                @click.stop=""
              >
                <v-icon class="ma-auto">mdi-near-me</v-icon>
              </v-list-item>
              <v-list-item
                v-else-if="!isMini"
                :to="{ name: 'states' }"
                variant="text"
                class="nav-states nav-regions"
                :ripple="false"
                @click.stop=""
              >
                <v-list-item-title class="nav-menu-item">
                  <v-icon>mdi-near-me</v-icon>
                  <p class="nav-item-title">
                    {{ $gettext(`Regions`) }}
                  </p>
                </v-list-item-title>
              </v-list-item>
            </template>

            <template v-if="canSearchPlaces">
              <v-list-item
                v-if="isMini"
                v-show="canSearchPlaces && $config.feature('places')"
                :to="{ name: 'places' }"
                variant="text"
                class="nav-places"
                :ripple="false"
                @click.stop=""
              >
                <v-icon class="ma-auto">mdi-map-marker</v-icon>
              </v-list-item>
              <div v-else v-show="canSearchPlaces && $config.feature('places')">
                <v-list-item to="/places" variant="text" class="nav-places activator" @click.stop="">
                  <v-list-item-title class="nav-menu-item">
                    <p class="nav-item-title">
                      {{ $gettext(`Places`) }}
                    </p>
                    <span v-show="config.count.places > 0" class="nav-count-group">{{ config.count.places }}</span>
                  </v-list-item-title>
                </v-list-item>

                <v-list-group>
                  <template #activator="{ props }">
                    <v-list-item v-bind="props" variant="text" class="nav-places" @click.stop="">
                      <v-icon>mdi-map-marker</v-icon>
                    </v-list-item>
                  </template>

                  <v-list-item :to="{ name: 'states' }" variant="text" class="nav-states nav-regions" @click.stop="">
                    <v-list-item-title :class="`nav-menu-item menu-item`">
                      {{ $gettext(`Regions`) }}
                    </v-list-item-title>
                    <span v-show="config.count.states > 0" class="nav-count-item">{{ config.count.states }}</span>
                  </v-list-item>
                </v-list-group>
              </div>
            </template>

            <v-list-item
              v-if="isMini && $config.feature('calendar')"
              :to="{ name: 'calendar' }"
              variant="text"
              class="nav-calendar"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-calendar</v-icon>
            </v-list-item>
            <v-list-item
              v-else-if="!isMini && $config.feature('calendar')"
              :to="{ name: 'calendar' }"
              variant="text"
              class="nav-calendar"
              :ripple="false"
              @click.stop=""
            >
              <v-list-item-title class="nav-menu-item">
                <v-icon>mdi-calendar</v-icon>
                <p class="nav-item-title">
                  {{ $gettext(`Calendar`) }}
                </p>
              </v-list-item-title>
              <span v-show="config.count.months > 0" class="nav-count-item">{{ config.count.months }}</span>
            </v-list-item>

            <v-list-item
              v-if="isMini && $config.feature('moments')"
              :to="{ name: 'moments' }"
              variant="text"
              class="nav-moments"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-filmstrip-box</v-icon>
            </v-list-item>
            <v-list-item
              v-else-if="!isMini && $config.feature('moments')"
              :to="{ name: 'moments' }"
              variant="text"
              class="nav-moments"
              :ripple="false"
              @click.stop=""
            >
              <v-list-item-title class="nav-menu-item">
                <v-icon>mdi-filmstrip-box</v-icon>
                <p class="nav-item-title">
                  {{ $gettext(`Moments`) }}
                </p>
              </v-list-item-title>
              <span v-show="config.count.moments > 0" class="nav-count-item">{{ config.count.moments }}</span>
            </v-list-item>

            <v-list-item
              v-if="isMini && $config.feature('labels')"
              to="/labels"
              variant="text"
              class="nav-labels"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-label</v-icon>
            </v-list-item>
            <v-list-item
              v-else-if="!isMini && $config.feature('labels')"
              to="/labels"
              variant="text"
              class="nav-labels"
              :ripple="false"
              @click.stop=""
            >
              <v-list-item-title class="nav-menu-item">
                <v-icon>mdi-label</v-icon>
                <p class="nav-item-title">
                  {{ $gettext(`Labels`) }}
                </p>
              </v-list-item-title>
              <span v-show="config.count.labels > 0" class="nav-count-item">{{ config.count.labels }}</span>
            </v-list-item>

            <v-list-item
              v-if="isMini && $config.feature('folders')"
              :to="{ name: 'folders' }"
              variant="text"
              class="nav-folders"
              :ripple="false"
              @click.stop=""
            >
              <v-icon class="ma-auto">mdi-folder</v-icon>
            </v-list-item>
            <v-list-item
              v-else-if="!isMini && $config.feature('folders')"
              :to="{ name: 'folders' }"
              variant="text"
              class="nav-folders"
              :ripple="false"
              @click.stop=""
            >
              <v-list-item-title class="nav-menu-item">
                <v-icon>mdi-folder</v-icon>
                <p class="nav-item-title">
                  {{ $gettext(`Folders`) }}
                </p>
              </v-list-item-title>
              <span v-show="config.count.folders > 0" class="nav-count-item">{{ config.count.folders }}</span>
            </v-list-item>

            <template v-if="!config.disable.settings">
              <v-list-item
                v-if="isMini"
                v-show="$config.feature('settings')"
                :to="{ name: 'settings' }"
                variant="text"
                class="nav-settings"
                :ripple="false"
                @click.stop=""
              >
                <v-icon class="ma-auto">mdi-cog</v-icon>
              </v-list-item>
            </template>

            <v-list-item v-show="!auth" :to="{ name: 'login' }" variant="text" class="nav-login" @click.stop="">
              <v-list-item-title>
                <v-icon>mdi-lock</v-icon>
                {{ $gettext(`Login`) }}
              </v-list-item-title>
            </v-list-item>

            <v-list-item
              v-if="isMini && featMembership"
              :to="{ name: 'upgrade' }"
              variant="text"
              class="nav-upgrade"
              @click.stop=""
            >
              <v-icon v-if="isPro" class="ma-auto">mdi-check-decagram</v-icon>
              <v-icon v-else class="ma-auto">mdi-diamond</v-icon>
            </v-list-item>
          </v-list>

          <div v-if="disconnected" class="nav-info connection-info clickable" @click.stop="showServerConnectionHelp">
            <div class="nav-info__underlay"></div>
            <div class="text-center">
              <v-icon icon="mdi-wifi-off" color="warning" size="21"></v-icon>
            </div>
            <div v-if="!isMini" class="text-start text-body-2">
              {{ $gettext(`No server connection`) }}
            </div>
          </div>
          <div v-else-if="!isMini && featUsage" class="nav-info usage-info clickable" @click.stop="showUsageInfo">
            <div class="nav-info__underlay"></div>
            <div class="nav-info__content">
              <v-progress-linear
                :model-value="config.usage.filesUsedPct"
                :color="config.usage.filesUsedPct > 95 ? 'error' : 'selected'"
                height="16"
                max="100"
                min="0"
                width="100%"
                rounded
              >
                <div class="text-caption opacity-85">
                  {{
                    $gettext(`%{n} GB of %{q} GB used`, {
                      n: $util.gigaBytes(config.usage.filesUsed),
                      q: $util.gigaBytes(config.usage.filesTotal),
                    })
                  }}
                </div>
              </v-progress-linear>
            </div>
          </div>

          <div v-show="auth && !isPublic && !disconnected" class="nav-info user-info">
            <div class="nav-info__underlay"></div>
            <div class="nav-user-avatar text-center my-1 mx-2 clickable" @click.stop="showAccountSettings">
              <img :src="userAvatarURL" :alt="accountInfo" :title="accountInfo" class="rounded-circle" />
            </div>
            <div v-if="!isMini" class="text-start mt-1 flex-grow-1 clickable" @click.stop="showAccountSettings">
              <p class="text-body-2">{{ displayName }}</p>
              <p class="text-caption opacity-70">{{ accountInfo }}</p>
            </div>
            <div class="text-center">
              <v-btn icon variant="text" :elevation="0" @click.stop.prevent="onLogout">
                <v-icon>mdi-power</v-icon>
              </v-btn>
            </div>
          </div>
          <div v-show="isPublic && !disconnected" class="nav-info user-info">
            <div class="nav-info__underlay"></div>
            <div class="text-start mt-1 flex-grow-1 clickable" @click.stop="showAccountSettings">
              <p class="text-body-2">Settings</p>
              <p class="text-caption opacity-70"></p>
            </div>
            <div class="text-center">
              <v-btn icon variant="text" :elevation="0" @click.stop.prevent="showAccountSettings">
                <v-icon>mdi-cog</v-icon>
              </v-btn>
            </div>
          </div>
        </div>
      </v-navigation-drawer>
      <div v-if="config.legalInfo" id="legal-info">
        <span v-if="config.legalUrl" class="clickable" @click.stop.prevent="showLegalInfo()">{{
          config.legalInfo
        }}</span>
        <span v-else>{{ config.legalInfo }}</span>
      </div>
    </template>

    <div id="mobile-menu" :class="{ active: speedDial }" @click.stop="speedDial = false">
      <div class="menu-content grow-top-end">
        <div class="menu-icons">
          <slot name="menu-icon">
            <a href="#" :title="$gettext('Reload')" class="menu-action nav-reload" @click.prevent="reloadApp">
              <v-icon>mdi-refresh</v-icon>
            </a>
            <router-link
              v-if="auth && $config.feature('settings') && !routeName('settings')"
              :to="{ name: 'settings' }"
              :title="$gettext('Settings')"
              class="menu-action nav-settings"
            >
              <v-icon>mdi-cog</v-icon>
            </router-link>
            <a
              v-if="auth && !config.readonly && $config.feature('upload')"
              href="#"
              :title="$gettext('Upload')"
              class="menu-action nav-upload"
              @click.prevent="openUpload()"
            >
              <v-icon>mdi-cloud-upload</v-icon>
            </a>
            <router-link
              v-if="!auth && !isPublic"
              :to="{ name: 'login' }"
              :title="$gettext('Login')"
              class="menu-action nav-login"
            >
              <v-icon>mdi-login</v-icon>
            </router-link>
          </slot>
        </div>
        <div class="menu-actions">

            <div v-if="auth && !routeName('browse') && $config.feature('search')" class="menu-action nav-search">
              <router-link to="/browse">
                <v-icon>mdi-view-comfy</v-icon>
                {{ $gettext(`Browse`) }}
              </router-link>
            </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import links from "common/links";

export default {
  name: "PNavigation",
  data() {
    const appName = this.$config.getName();

    let appNameSuffix = "";
    let appNameParts = appName.split(" ");

    if (appNameParts.length > 1) {
      appNameSuffix = appNameParts.slice(1, 9).join(" ");
    }

    const canManagePhotos = this.$config.allow("photos", "manage");
    const isDemo = this.$config.get("demo");
    const isPro = !!this.$config.values?.ext["pro"];
    const isPublic = this.$config.get("public");
    const isReadOnly = this.$config.get("readonly");
    const isRestricted = this.$config.deny("photos", "access_library");
    const isSuperAdmin = this.$session.isSuperAdmin();
    const tier = this.$config.getTier();

    return {
      links,
      canSearchPlaces: this.$config.allow("places", "search"),
      canAccessPrivate: !isRestricted && this.$config.allow("photos", "access_private"),
      canManagePhotos: canManagePhotos,
      canManagePeople: this.$config.allow("people", "manage"),
      canManageUsers: (!isPublic || isDemo) && this.$config.allow("users", "manage"),
      appNameSuffix: appNameSuffix,
      appName: this.$config.getName(),
      appAbout: this.$config.getAbout(),
      appIcon: this.$config.getIcon(),
      siteUrl: this.$config.values.siteUrl,
      disconnected: this.$config.disconnected,
      indexing: false,
      drawer: null,
      featUpgrade: tier < 6 && isSuperAdmin && !isPublic && !isDemo,
      featMembership: tier < 3 && isSuperAdmin && !isPublic && !isDemo,
      featFeedback: tier >= 6 && isSuperAdmin && !isPublic && !isDemo,
      featFiles: this.$config.feature("files"),
      featUsage: canManagePhotos && this.$config.feature("files") && this.$config.values?.usage?.filesTotal,
      isRestricted: isRestricted,
      isMini: localStorage.getItem("navigation.mode") !== "false" || isRestricted,
      isDemo: isDemo,
      isPro: isPro,
      isPublic: isPublic,
      isReadOnly: isReadOnly,
      isAdmin: this.$session.isAdmin(),
      isSuperAdmin: isSuperAdmin,
      isSponsor: this.$config.isSponsor(),
      isTest: this.$config.test,
      session: this.$session,
      config: this.$config.values,
      page: this.$config.page,
      user: this.$session.getUser(),
      speedDial: false,
      rtl: this.$isRtl,
      subscriptions: [],
    };
  },
  computed: {
    auth() {
      return this.session.auth || this.isPublic;
    },
    visible() {
      return !this.$route.meta.hideNav;
    },
    displayName() {
      const user = this.$session.getUser();
      if (user) {
        return user.getDisplayName();
      }

      return this.$gettext("Unregistered");
    },
    userAvatarURL() {
      return this.$session.getUser().getAvatarURL("tile_50", this.$config);
    },
    accountInfo() {
      const user = this.$session.getUser();
      if (user) {
        return user.getAccountInfo();
      }

      return this.$gettext("Account");
    },
  },
  created() {
    this.subscriptions.push(this.$event.subscribe("index", this.onIndex));
    this.subscriptions.push(this.$event.subscribe("import", this.onIndex));
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    routeName(name) {
      if (!name || !this.$route.name) {
        return false;
      }

      return this.$route.name.startsWith(name);
    },
    reloadApp() {
      this.$notify.info(this.$gettext("Reloading…"));
      this.$notify.blockUI();
      setTimeout(() => window.location.reload(), 100);
    },
    openUpload() {
      this.$event.publish("dialog.upload");
    },
    goHome() {
      if (this.$route.name !== "home") {
        this.$router.push({ name: "home" });
      }
    },
    showDrawer() {
      if (this.auth) {
        this.drawer = true;
        this.isMini = this.isRestricted;
      }
    },
    hideDrawer() {
      if (this.auth) {
        this.drawer = false;
        this.isMini = this.isRestricted;
      }
    },
    toggleDrawer(ev) {
      ev?.preventDefault();
      ev?.stopPropagation();

      if (this.$vuetify.display.smAndDown) {
        if (this.drawer) {
          this.hideDrawer();
        } else {
          this.showDrawer();
        }
      } else {
        this.toggleIsMini();
      }
    },
    toggleIsMini() {
      if (this.isRestricted) {
        return;
      }

      this.isMini = !this.isMini;
      localStorage.setItem("navigation.mode", `${this.isMini}`);
    },
    showAccountSettings() {
      if (this.$config.feature("account")) {
        this.$router.push({ name: "settings_account" });
      } else {
        this.$router.push({ name: "settings" });
      }
    },
    showUsageInfo() {
      this.$router.push({ path: "/index/files" });
    },
    showServerConnectionHelp() {
      this.$router.push({ path: "/help/websockets" });
    },
    showLegalInfo() {
      if (this.config.legalUrl) {
        window.open(this.config.legalUrl, "_blank").focus();
      } else {
        this.$router.push({ name: "about" });
      }
    },
    onLogout() {
      this.$session.logout();
    },
    onIndex(ev) {
      if (!ev) {
        return;
      }

      const type = ev.split(".")[1];

      switch (type) {
        case "file":
        case "folder":
        case "indexing":
          this.indexing = true;
          break;
        case "completed":
          this.indexing = false;
          break;
      }
    },
  },
};
</script>
