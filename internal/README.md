# internal

В этой директории размещается код внутренних модулей приложения. Код внутри этого пакета недоступен для импорта в других приложениях.

Директория `internal/` является специальной в Go и обеспечивает инкапсуляцию кода на уровне модуля. Компилятор Go запрещает импорт пакетов из `internal/` за пределами родительского модуля.

export "host=%s user=%s password=%s dbname=%s sslmode=disable",
`localhost`, `video`, `XXXXXXXX`, `video`

host=host port=port user=myuser password=xxxx dbname=mydb sslmode=disable