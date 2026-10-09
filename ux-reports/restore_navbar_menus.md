# UX — Restauració de Menús a la Barra de Navegació (`AppNavbar.vue`)

**Veredicte**: APTE  
**Data**: 2026-10-09  

## Flux revisats
- **Navegació com a Professor / Docent**: Comprovació de l'accés complet als mòduls de docència (Jocs & Quizzes, Avaluacions, Grups, Usuaris, Cursos, Materials i Manual del Professor).
- **Navegació com a Alumne (Estudiant)**: Validació de la vista simplificada i segura, visualitzant únicament els mòduls permesos (Jocs & Quizzes, Grups, Cursos i Materials) i ocultant panells docents/administratius (Avaluacions, Usuaris, Manual i Mètriques).
- **Navegació com a Administrador**: Verificació de la disponibilitat global de tots els menús incloent el panell de control de Mètriques & Auditoria.
- **Canvi d'idioma dinàmic**: Comprovació de la reactivitat instantània de les etiquetes de text de la barra de navegació en català (CA), castellà (ES) i anglès (EN).
- **Adaptabilitat responsive mòbil (<=768px)**: Validació de la reducció de les etiquetes de text mantenint la iconografia neta i funcional per a pantalles tàctils.

## Usabilitat
- [OK] **Accés complet i coherent als recursos**: La recuperació dels accessos directes a *Cursos*, *Materials* i *Usuaris* proporciona una navegació fluida i sense passos intermedis innecessaris cap als continguts lectius i la gestió de l'alumnat.
- [OK] **Separació neta de privilegis**: Els usuaris amb rol d'alumne disposen d'un menú clar i sense sobrecàrrega visual ni elements inaccessibles deshabilitats.
- [OK] **Drecera de perfil i tancament de sessió**: La integració de l'accés directe al perfil mitjançant l'avatar/nom i el botó dedicat de tancament de sessió asseguren una gestió d'usuari intuïtiva.
- [OK] **Identificació de rol**: El badge/etiqueta de rol (`Tag`) permet a l'usuari saber en tot moment el seu nivell d'accés (Admin, Professor, Alumne).

## Consistència visual
- [OK] **Iconografia semàntica coherent**: Ús de PrimeIcons altament reconeixibles i representatius per a cada mòdul (`pi-th-large` per a Quizzes, `pi-chart-bar` per a Avaluacions, `pi-users` per a Grups, `pi-user-plus` per a Usuaris, `pi-book` per a Cursos, `pi-folder-open` per a Materials, `pi-question-circle` per al Manual i `pi-chart-line` per a Mètriques).
- [OK] **Components PrimeVue**: Ús impecable de `Tag` i `Button` amb severitats visuals estandarditzades (`danger`, `info`, `success`, `secondary`).
- [OK] **Paleta de colors i capçalera fixa**: Fons blanc amb vora inferior subtil `#e2e8f0`, ombra lleugera i comportament `sticky` a la part superior que garanteix la disponibilitat permanent de la navegació.

## Responsive/mòbil
- [OK] **Mode compacte a mòbil (<=768px)**: Ocultació automàtica del text de les etiquetes (`.nav-link span`) i dels noms llargs d'usuari/marca, mostrant exclusivament la iconografia essencial.
- [OK] **Dimensions tàctils adequades**: Els botons i enllaços mantenen una àrea tàctil còmoda i espaiada (`gap: 0.5rem`, padding `0.5rem 0.85rem`), evitant pulsacions accidentals.
- [OK] **Selector d'idioma adaptable**: Mida reduïda i compacta amb icona de globus terràqüi que cap perfectament a la capçalera mòbil sense provocar desbordament horitzontal.

## Feedback a l'usuari
- [OK] **Estat de ruta activa ressaltat**: L'element de navegació corresponent a la pàgina actual es destaca clarament amb fons índigo clar `#e0e7ff` i color primari d'alta intensitat `#4338ca` amb pes tipogràfic `600`.
- [OK] **Transicions i estats hover**: Resposta visual immediata en passar el cursor per sobre de qualsevol enllaç, botó d'idioma o botó de desconnexió.
- [OK] **Canvi d'idioma reactiu**: Actualització instantània de totes les etiquetes de la barra en commutar d'idioma sense refrescar la pàgina.

## Accessibilitat bàsica
- [OK] **Cobertura i18n total**: Claus definides de manera homogènia a `ca.json`, `es.json` i `en.json` (`nav.quizzes`, `nav.evaluations`, `nav.groups`, `nav.users`, `nav.courses`, `nav.materials`, `nav.manual`, `nav.metrics`, `nav.roles.*`).
- [OK] **Contrast de color contrastat**: Contrast cromàtic òptim tant en estat normal (`#475569` sobre `#ffffff`) com en actiu (`#4338ca` sobre `#e0e7ff`).
- [OK] **Atributs descriptius**: Inclusió d'atributs `title` al selector d'idioma, enllaç de perfil i botó de logout per facilitar la navegació a lectors de pantalla i proporcionar tooltips natius.

## Incidències
1. **[Detall] Desplegable tipus 'Hamburguesa' per a resolucions molt petites (<480px)**: A pantalles de mòbil especialment estretes (<375px) amb perfils d'administrador (on conflueixen 8 icones més selector d'idioma i perfil), podria ser convenient en futures versions implementar un menú col·lapsable lateral o desplegable inferior per millorar encara més la folgança de l'espai.
