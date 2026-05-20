<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
    <div class="w-full max-w-md bg-white rounded-xl shadow-lg p-8">
      <h1 class="text-2xl font-bold text-center text-gray-800 mb-2">Вход в TaskEngine</h1>
      <p class="text-center text-gray-500 mb-6">Введите данные для доступа к проектам</p>

      <Form v-slot="{ errors, isSubmitting }" :validation-schema="schema" @submit="onSubmit" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <Field name="email" type="email" class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" placeholder="you@company.com" />
          <p v-if="errors.email" class="mt-1 text-xs text-red-500">{{ errors.email }}</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
          <Field name="password" type="password" class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" placeholder="••••••••" />
          <p v-if="errors.password" class="mt-1 text-xs text-red-500">{{ errors.password }}</p>
        </div>

        <div class="flex items-center justify-between text-sm">
          <label class="flex items-center gap-2 cursor-pointer select-none">
            <input type="checkbox" class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500" />
            <span class="text-gray-600">Запомнить меня</span>
          </label>
          <RouterLink to="/register" class="text-blue-600 hover:underline font-medium">Нет аккаунта?</RouterLink>
        </div>

        <button type="submit" :disabled="isSubmitting" class="w-full bg-blue-600 text-white py-2.5 rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors shadow-sm">
          {{ isSubmitting ? 'Вход...' : 'Войти' }}
        </button>

        <p v-if="formError" class="text-center text-sm text-red-600 bg-red-50 p-2 rounded-lg">{{ formError }}</p>
      </Form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Form, Field } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'
import { useAuth } from '@/composables/useAuth'

const { login } = useAuth()
const formError = ref('')

const schema = toTypedSchema(z.object({
  email: z.string().email('Некорректный email'),
  password: z.string().min(6, 'Минимум 6 символов'),
}))

const onSubmit = async (values: any, { setErrors }: any) => {
  formError.value = ''
  try {
    await login(values.email, values.password)
  } catch (err: any) {
    formError.value = err.response?.data?.message || 'Ошибка авторизации. Проверьте данные.'
  }
}
</script>
