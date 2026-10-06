# UX — Manual de Professor i Canvi d'Idioma (i18n)

**Veredicte**: APTE  
**Data**: 2026-10-06  

## Flux revisats
- **Consulta del Manual del Professorat (`TeacherManualView.vue`)**: Navegació detallada per les 10 seccions del manual (Introducció i Filosofia, Autenticació, Grups, Alumnes, Quizzes, Cursos, Materials, Guió de Classe, Partida en Directe i Panell d'Avaluació).
- **Canvi d'idioma en temps real (`AppNavbar.vue` & `useI18n`)**: Selecció entre Català (CA), Castellà (ES) i Anglès (EN) i validació de la reactivitat immediata a tots els blocs de text, capçaleres, acordions i botons de la interfície.
- **Navegació per seccions i desplaçament suau (`scrollToSection`)**: Ús de l'índex lateral sticky per saltar directament a cada secció amb scroll animat i marge superior configurat per a la barra de navegació.
- **Visualització adaptativa en mòbil i tauleta**: Prova de la distribució visual en diferents amples de pantalla (>=1280px, <=992px, <=768px).

## Usabilitat
- [OK] **Estructura clara i pedagògica**: Organització en 10 seccions temàtiques ben diferenciades, amb icones descriptives per a cada mòdul d'Encertia.
- [OK] **Diversitat de patrons visuals d'aprenentatge**: Ús de targetes explicatives (`features-grid`), llistes de passos numerats per a fluxos complexos (alta d'alumnes, partides), acordions per a temes tècnics (tokens JWT, materials) i targetes comparatives (punts de joc vs. nota acadèmica).
- [OK] **Llegibilitat del contingut**: Encapçalaments clars, mida de lletra adequada i espaiat consistent que faciliten la lectura a professors no tècnics.
- [Millorable] **Actualització de l'índex en scroll manual**: L'índex lateral destaca la secció activa en clicar-hi, però no actualitza l'estat d'activació quan l'usuari es desplaça manualment fent scroll per la pàgina.

## Consistència visual
- [OK] **Ús coherent de PrimeVue**: Integració d'elements de PrimeVue (`Card`, `Accordion`, `AccordionPanel`, `AccordionHeader`, `AccordionContent`, `Tag`, `Message`) en consonància amb la resta de mòduls de la plataforma.
- [OK] **Paleta de colors i iconografia**: Ús del color primari d'Encertia (`#4f46e5`), tons pastel neutres per als fons de contingut i etiquetes de severitat (`Tag`) alineades amb la barra de navegació superior.
- [OK] **Tipografia i Targetes**: Homogeneïtat en les ombres, vora de 1px (`#e2e8f0`) i radi de curvatura (`0.75rem`), aconseguint un aspecte polit i professional.

## Responsive/mòbil
- [OK] **Reorganització de graelles**: Les funcions (`.features-grid`) i blocs de guió (`.block-types-container`) utilitzen `auto-fit` amb amplada mínima de 220px, adaptant-se automàticament a pantalles petites.
- [OK] **Comparativa de notes a una columna**: En pantalles petites (<=992px), la comparació entre punts de joc i nota acadèmica es mostra en una sola columna per evitar compressió de text.
- [Millorable] **Navegació mòbil sense índex alternatiu**: A pantalles menors de 992px s'oculta la barra lateral (`display: none`). Encara que manté la vista neta, seria recomanable incloure un menú desplegable superior o botó flotant d'índex per a navegació ràpida en mòbil.

## Feedback a l'usuari
- [OK] **Estat actiu clar a l'índex**: L'enllaç de la secció seleccionada es ressalta amb fons blau clar (`#eef2ff`) i text en negreta primària (`#4f46e5`).
- [OK] **Canvi d'idioma instantani**: La selecció de llengua des de la navbar actualitza immediatament totes les seccions del manual sense necessitat de recarregar la pàgina.
- [OK] **Avisos destacats per a conceptes clau**: Utilització de components `Message` (severitat `info` i `success`) per ressaltar l'estil visual Kahoot i les indicacions de control en directe del professor.

## Accessibilitat bàsica
- [OK] **Coberta i18n del 100%**: Les tres llengües oficials d'Encertia (CA, ES, EN) tenen una correspondència exacta de claus 1:1 als fitxers `ca.json`, `es.json` i `en.json`, sense cap clau o text en dur desemparat.
- [OK] **Contrast i jerarquia HTML**: Alt contrast de text sobre fons blanc i capçalera principal amb gradient de gran llegibilitat. Ús d'etiquetes semàntiques HTML5 (`<header>`, `<aside>`, `<nav>`, `<main>`, `<section>`).

## Incidències
1. **[Detall] Sync d'Active Section en Scroll**: En navegar amb la roda del ratolí o deslizable tàctil, la barra lateral no actualitza automàticament quina secció s'està llegint. Es suggereix connectar un `IntersectionObserver` a `TeacherManualView.vue`.
2. **[Detall] Índex mòbil alternatiu**: Quan la pantalla és <=992px, la barra lateral s'amaga. S'aconsella afegir un dropdown o Drawer col·lapsable per facilitar els salts de secció des del mòbil.
