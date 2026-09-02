# Encertia — Document Funcional

Plataforma d'avaluació i gestió de curs per a docència universitària.

> Aquest document defineix **què** ha de fer la plataforma (abast funcional).
> No conté detalls tècnics d'implementació (stack, arquitectura, contractes
> d'API) — per això veure `constitution.md` i `contracts/*.openapi.yaml`.
> Tampoc conté calendari ni condicions econòmiques.

## 1. Introducció i objectiu

Encertia combina un sistema de qüestionaris interactius en directe (a
l'estil Kahoot) amb un registre persistent de respostes per alumne i una
gestió de curs completa — material, unitats didàctiques i seguiment
individual — perquè el professorat pugui avaluar el rendiment més enllà
del resultat immediat del joc.

## 2. Usuaris i rols

- **Professor** — gestiona el curs, puja material, crea qüestionaris,
  dona d'alta als alumnes, llança partides en directe i avalua els
  resultats a posteriori.
- **Alumne** — s'uneix a les partides amb el seu compte, respon les
  preguntes en directe, consulta material i el seu propi progrés.

Els comptes d'alumne els crea el professor (individualment o important una
llista); no hi ha auto-registre obert, per mantenir el control del
grup-classe.

## 3. Abast funcional

### 3.1 Autenticació i comptes

Accés diferenciat per professor i alumne, amb sessió segura.

- Login amb email i contrasenya per a professor i alumnes
- Missatges d'error clars sense revelar informació sensible sobre els
  comptes (no s'indica si l'error és l'email o la contrasenya)
- Sessió amb caducitat automàtica per seguretat (24 hores d'inactivitat)
- Tancament de sessió efectiu (logout real al servidor, no només al
  navegador)
- Permisos diferenciats: un alumne mai pot accedir a funcions de gestió
  del professor

### 3.2 Gestió d'alumnes

Alta i manteniment del llistat d'alumnes del curs.

- Alta individual d'alumnes (nom, email, i qualsevol identificador
  acadèmic que calgui)
- Alta massiva important una llista (per exemple, un full de càlcul amb
  tot el grup-classe)
- Edició de dades d'un alumne ja donat d'alta
- Baixa o desactivació d'un alumne sense esborrar el seu historial de
  resultats
- Llistat filtrable i cercable del grup-classe

### 3.3 Gestió de curs i unitats didàctiques

Estructura del curs en unitats seqüencials. **"Unitat didàctica" i "classe"
són el mateix concepte** — es fan servir indistintament segons el context
(acadèmic o de sessió lectiva), sense que hi hagi cap nivell addicional a
l'estructura.

- Creació d'un curs amb nom, descripció i dates d'inici/fi
- Organització en unitats/temes, en l'ordre que decideixi el professor
- **Cada unitat pot vincular-se a diversos qüestionaris/jocs (relació N a
  N, no 1 a 1)** — per exemple, un joc de repàs a meitat d'unitat i un
  altre d'avaluació final per la mateixa unitat
- Cada unitat pot tenir el seu propi material (documents, vídeo)
  independent de les altres
- Seguiment visual de progrés de l'alumne unitat per unitat
  (completat/pendent)

### 3.4 Material didàctic

Documents i vídeo associats a cada unitat.

- Pujada de documents (PDF, Word) descarregables per l'alumne
- **Els documents PDF es poden visualitzar directament dins l'aplicació**
  (visor integrat, pàgina a pàgina), no només descarregar-se — necessari
  per al guió de classe (veure 3.6)
- Vídeo incrustat dins la plataforma (allotjat a un servei extern tipus
  Vimeo/YouTube no llistat, per no dependre d'emmagatzematge propi)
- Un mateix material es pot reutilitzar en més d'una unitat si cal
- Substitució d'un document per una versió actualitzada sense trencar
  l'enllaç que ja tenen els alumnes
- Visibilitat, per al professor, de qui ha consultat cada material i quan

### 3.5 Creació de qüestionaris

Editor de preguntes i qüestionaris reutilitzables.

- **Tipus de pregunta suportats**: opció múltiple, veritable/fals, i
  ordenació (l'alumne ha d'ordenar una sèrie d'elements correctament)
- Temps de resposta i puntuació configurables per pregunta,
  independentment del tipus
- Agrupació de preguntes en qüestionaris reutilitzables amb nom propi
- Duplicació d'un qüestionari existent com a punt de partida per crear-ne
  un de nou
- **Un mateix qüestionari es pot vincular a més d'una unitat didàctica**
  si el professor ho vol (mateixa relació N a N que a 3.3)
- Banc de preguntes reutilitzable entre qüestionaris diferents, per no
  haver de redactar-les cada cop, independentment del tipus de pregunta

### 3.6 Guió de classe (visor seqüencial)

Dins de cada unitat didàctica, el professor pot dissenyar per endavant un
**guió reproduïble** que combina material i qüestionaris en l'ordre exacte
que vol impartir la classe — pensat per fer-lo servir en directe durant la
sessió real, sense haver de canviar d'aplicació.

- El guió és una **seqüència ordenada de blocs**, de tres tipus possibles:
  - **Bloc de material**: un rang concret de pàgines d'un PDF/document ja
    pujat (per exemple, "diapositives 1 a 5"), mostrat amb el visor
    integrat (veure 3.4)
  - **Bloc de qüestionari**: llança automàticament una partida en directe
    d'un qüestionari concret ja creat (veure 3.5)
  - **Bloc de pausa/torn de preguntes**: una marca sense contingut
    específic, només perquè el guió recordi al professor que toca aturar-se
    per preguntes obertes del grup
- Durant la classe real, el professor **reprodueix el guió** amb un botó
  d'avançar (Play/Següent), que va passant seqüencialment pels blocs
- Dins d'un bloc de material, avançar mostra la pàgina següent del rang
  definit; en arribar al final del rang, el següent clic passa
  automàticament al bloc següent del guió
- Dins d'un bloc de qüestionari, avançar llança la partida en directe
  corresponent (codi/PIN, preguntes, rànquing — mòdul 3.7)
- Es pot pausar, retrocedir o saltar blocs manualment en qualsevol moment,
  per si la classe real no segueix exactament el pla previst
- Un guió es pot duplicar com a punt de partida per crear-ne un altre de
  similar (per exemple, per un altre grup del mateix curs)

**Aclariment d'abast**: el visor és d'ús exclusiu del professor — es
projecta a l'aula (pantalla/projector), com faria amb qualsevol
presentació. No cal sincronització en temps real amb els dispositius dels
alumnes per als blocs de material. L'única part que sí es sincronitza amb
els alumnes és el bloc de qüestionari, que reutilitza el mateix mecanisme
de partida en directe ja descrit a 3.7 (codi/PIN, connexió des del mòbil).
L'objectiu del guió és mantenir l'atenció de la classe alternant
explicació i moments interactius, no crear una experiència multi-pantalla
sincronitzada.

### 3.7 Partida en directe

L'experiència de joc en temps real, a l'estil Kahoot.

- Codi/PIN generat per cada partida perquè els alumnes s'hi uneixin des
  del mòbil o ordinador
- El professor controla el ritme: llança cada pregunta quan el grup està
  a punt
- Preguntes mostrades en directe amb temporitzador visible per tothom
- Rànquing en temps real després de cada pregunta, calculat amb els
  **punts de joc** (veure 3.8): com més ràpida i correcta la resposta,
  més punts — amb podi final a l'estil Kahoot
- Reconnexió sense pèrdua de dades si un alumne perd la connexió a mitja
  partida
- Cada resposta queda registrada a l'instant (resposta donada, correcta
  o no, i temps trigat), associada a l'alumne — no és anònima com en un
  Kahoot convencional

### 3.8 Sistema de doble puntuació: joc vs. avaluació

Cada resposta es guarda una sola vegada (resposta donada, si és correcta,
i el temps trigat), però **es fan servir dues escales de puntuació
diferents** segons el propòsit:

- **Punts de joc** (gamificació, visibles durant la partida en directe):
  tenen en compte tant l'encert com la rapidesa — respondre bé i ràpid
  dona més punts que respondre bé i lent. És el que alimenta el rànquing
  i el podi en directe.
- **Nota d'avaluació** (acadèmica, visible al panell d'avaluació
  posterior): **només té en compte si la resposta és correcta o no**. El
  temps de resposta queda registrat i disponible per consulta, però no
  afecta la nota de cap manera.

Aquesta separació permet que la partida en directe mantingui l'emoció i
el ritme d'un joc, sense que la pressió del temps distorsioni la nota
acadèmica real de l'alumne.

### 3.9 Avaluació i qualificació

El que diferencia Encertia d'un Kahoot convencional.

- Resultats detallats per alumne de cada partida: quines preguntes ha
  encertat, fallat o deixat sense respondre, i en quant de temps (el
  temps es mostra com a dada informativa, sense afectar la nota — veure 3.8)
- Possibilitat de revisar i corregir manualment una resposta puntuada
  automàticament pel sistema (per exemple, si una pregunta tenia més
  d'una resposta vàlida)
- Exportació de resultats (Excel/CSV) per incorporar-los a l'acta o
  expedient acadèmic
- Estadístiques per pregunta: percentatge d'encert, temps mitjà de
  resposta, preguntes amb més dificultat per detectar conceptes mal
  entesos pel grup
- Nota consolidada per alumne combinant tots els qüestionaris de totes
  les unitats del curs, amb el detall de cada partida disponible per si
  cal revisar-la
- **Opció configurable per excloure les dues pitjors notes**: el
  professor pot activar un ajust que, en calcular la mitjana d'un alumne
  (a nivell de curs o de tema/unitat), descarti automàticament les dues
  notes més baixes obtingudes. Desactivat per defecte; quan s'activa,
  s'aplica de manera consistent a tot el càlcul de mitjanes corresponent

## 4. Dades i privacitat

Les dades dels alumnes (respostes, resultats, progrés) són propietat del
professor/institució, allotjades en una base de dades pròpia del
projecte — no dependents d'un tercer com Kahoot. Cap resposta ni resultat
es perd mai: les operacions d'esborrat sobre dades acadèmiques queden
desactivades per defecte (es marquen com a inactives, no s'eliminen
físicament), per preservar l'historial d'avaluació.

## 5. Relació amb altres documents del projecte

- `roadmap.md` — ordre previst d'implementació dels mòduls que cobreixen
  aquest abast funcional.
- `constitution.md` — regles tècniques i d'arquitectura (stack, estructura
  de carpetes, rols d'agent).
- `contracts/*.openapi.yaml` — contracte tècnic mòdul a mòdul, ja validat,
  que concreta com s'implementa cada punt d'aquest document.

Quan hi hagi contradicció entre aquest document i un contracte ja validat,
**aquest document defineix la intenció de producte; el contracte defineix
com s'implementa**. Si un canvi aquí implica trencar un contracte ja
validat, cal passar-ho per l'Orquestrador i validació humana (veure
`constitution.md`, secció 5).
