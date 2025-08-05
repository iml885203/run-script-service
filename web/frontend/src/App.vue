<template>
  <div id="app">
    <nav class="navbar">
      <div class="nav-brand">
        <h1>Run Script Service</h1>
      </div>
      <div class="nav-links">
        <router-link to="/" class="nav-link" active-class="active">
          Dashboard
        </router-link>
        <router-link to="/scripts" class="nav-link" active-class="active">
          Scripts
        </router-link>
        <router-link to="/logs" class="nav-link" active-class="active">
          Logs
        </router-link>
        <router-link to="/settings" class="nav-link" active-class="active">
          Settings
        </router-link>
      </div>
    </nav>

    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup lang="ts">
// No additional logic needed for the root component
</script>

<style scoped>
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background: var(--color-background-soft);
  border-bottom: 1px solid var(--color-border);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.nav-brand h1 {
  margin: 0;
  color: var(--color-text);
  font-size: 1.5rem;
  font-weight: 600;
}

.nav-links {
  display: flex;
  gap: 2rem;
}

.nav-link {
  text-decoration: none;
  color: var(--color-text-muted);
  font-weight: 500;
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  transition: all 0.2s ease;
}

.nav-link:hover {
  color: var(--color-text);
  background: var(--color-background-mute);
}

.nav-link.active {
  color: var(--color-brand);
  background: var(--color-brand-soft);
}

.main-content {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .navbar {
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
  }

  .nav-links {
    gap: 1rem;
  }

  .main-content {
    padding: 1rem;
  }
}
</style>
