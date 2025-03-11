# Используем официальный образ Node.js
FROM node:alpine

# Устанавливаем рабочую директорию в контейнере
WORKDIR /app

# Копируем package.json и package-lock.json для установки зависимостей
COPY package*.json ./

# Устанавливаем все зависимости
RUN npm install

# Копируем весь проект в контейнер
COPY ./frontend/dist /app

# Собираем проект
RUN npm run build

# Открываем порт, на котором будет работать React
EXPOSE 5173

# Запускаем приложение
CMD ["npm", "run", "dev"]
