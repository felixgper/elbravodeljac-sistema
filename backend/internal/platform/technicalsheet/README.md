# Ficha técnica original

Go incluye assets/pagina-1.jpg, pagina-2.jpg y catalogo.json. Fondos de 300 ppp
saneados del PDF de dos páginas proporcionado por el usuario. Se retiraron píxeles
Y texto OCR de muestras de cliente, documento, contacto, vehículo, operario, fechas,
trabajos, importes y selecciones. No se incorpora el PDF original con datos personales.
El logo, diseño, etiquetas y casillas originales permanecen. RUC y dirección se
completan desde PostgreSQL. Fuente SHA-256: 1c51f6c1df4d63ef52b22ee8a37c1138408d3ffc9e3c4d6e17c156615515592b.

La preparación se verificó con PyMuPDF/Poppler en desarrollo; no son dependencias de
runtime. El generador Go produce dos páginas A4, fondos raster y texto dinámico vectorial,
con espacios originales de conformidad. Toda selección empieza sin evaluar.

Los límites son explícitos: no se recorta contenido para fingir éxito, no se añaden
páginas ni se inventan firmas. Render devuelve error si el contenido no cabe de forma
legible o si la tipografía Windows-1252 no lo representa. La capa PostgreSQL conserva
el snapshot completo y la entrega, y el endpoint informa el error de exportación.
