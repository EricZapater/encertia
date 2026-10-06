# UX — Mòdul group

**Veredicte**: A MILLORAR  
**Data**: 2026-10-06  

## Flux revisats
- **Professor/Admin crea un nou grup**: Formulari molt accessible amb nom obligatori i el curs acadèmic inicialitzat per defecte a "26-27" amb el seu corresponent placeholder.
- **Llistat i filtrat de grups**: Taula interactiva amb cerca en temps real (debounce de 350ms), filtre per curs acadèmic, badges visuals i paginació lazy.
- **Assignació i desassignació d'alumnes**: Modal dedicat per afegir alumnes (mitjançant llista de UUIDs) i llistat de membres assignats amb acció de desvinculació individual.
- **Baixa lògica de grup (soft-delete)**: Modal de confirmació d'esborrat lògic amb indicació clara sobre la preservació de la informació acadèmica dels alumnes.

## Usabilitat
- [OK] Formulari de creació ràpid i senzill. El camp de curs acadèmic suggereix i preomple la proposta "26-27" per defecte segons els requisits d'HU-GROUP-01.
- [Millorable] De demanar UUIDs manuals en text pla per vincular un curs o afegir alumnes pot ser un punt de fricció per a professors no tècnics. Es recomana oferir desplegables o cercadors integrats.

## Consistència visual
- [OK] Ús impecable dels components PrimeVue (`DataTable`, `Column`, `Dialog`, `Button`, `InputText`, `useToast`) mantenint l'estètica global de l'aplicació.
- [Millorable] **Integració amb i18n incompleta**: La vista `GroupsListView.vue` té tots els textos de la interfície en línia (hardcoded) en català en lloc d'utilitzar les claus de la biblioteca d'i18n.

## Responsive/mòbil
- [OK] La vista s'adapta correctament a pantalles mòbils (`@media (max-width: 768px)`), reordenant la capçalera i la barra de filtres en columna.
- [OK] Diàlegs modals adaptats a `width: 90vw` i la taula de dades permet scroll horitzontal sense deformar la interfície.

## Feedback a l'usuari
- [OK] Notificacions d'èxit i error integrades mitjançant `useToast` de PrimeVue per a totes les operacions de creació, edició, baixa i assignació.
- [OK] Indicadors visuals de carregament (`:loading`) als botons durant la desada i estat buit descriptiu (`#empty`) a la taula.

## Accessibilitat bàsica
- [OK] Bon contrast de colors (utilitzant paleta slate/indigo), definició correcta d'etiquetes `<label>` amb `for` enllaçat a l'input i tooltips explicatius en tots els botons d'acció.

## Incidències
1. **[Notable] Mancança de traduccions i18n al mòdul de grups**: Tots els textos de la interfície a `GroupsListView.vue` (títol, headers de taula, botons, labels de formulari) estan directament en català. Cal afegir la secció `"groups"` als fitxers de traducció (`ca.json`, `es.json` i `en.json`) i substituir-los per la funció `t()`.
2. **[Detall] Entrada manual de UUIDs**: Tant al camp "UUID Curs Associat" com a la secció "Assignar nous alumnes", es demana enganxar UUIDs en format text. Per millorar l'experiència d'usuari del professorat, seria idoni utilitzar selectors desplegables (`Dropdown` / `MultiSelect`) amb les dades dels cursos i alumnes de la plataforma.
