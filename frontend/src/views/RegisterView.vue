<template>
  <div class="register-view">
    <h1>Регистрация</h1>

    <form @submit.prevent="handleRegister">
      <input
        v-model="email"
        type="email"
        placeholder="Email"
        required
      />
      <input
        v-model="password"
        type="password"
        placeholder="Пароль"
        required
      />
      <button type="submit">Зарегистрироваться</button>
    </form>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="success" class="success">Успешно! Перенаправление...</p>
    <router-link to="/login">Уже есть аккаунт? Войти</router-link>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const success = ref(false)

const handleRegister = async () => {
  try {
    await authStore.register(email.value, password.value)
    success.value = true
    setTimeout(() => router.push('/login'), 2000)
  } catch (err) {
    error.value = err.response?.data?.error || 'Ошибка регистрации'
  }
}
</script>

<style scoped>
.register-view {
  max-width: 400px;
  margin: 50px auto;
  padding: 20px;
}
input {
  display: block;
  width: 100%;
  margin: 10px 0;
  padding: 10px;
}
button {
  width: 100%;
  padding: 10px;
  background: #28a745;
  color: white;
  border: none;
  border-radius: 4px;
}
.error { color: red; }
.success { color: green; }
</style>
