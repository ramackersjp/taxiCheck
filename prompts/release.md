# Skill: Release V2.0.1

## Doel

Maak een volledige release van versie **V2.0.1** op basis van de huidige repository en de hieronder beschreven release-eisen.

De release moet volledig aansluiten op de huidige `dev`-branch. De nieuwe releaseversie en `dev` moeten inhoudelijk en qua versienummering correct op elkaar aansluiten.

## Releaseversie

- Releaseversie: `V2.0.1`
- Gebruik overal consequent dezelfde versie: `2.0.1` waar tooling een numerieke versie verwacht en `V2.0.1` waar de projectconventie de `V` gebruikt.
- Controleer de volledige repository op oude, afwijkende of inconsistente versienummers.
- Controleer daarbij ook documentatie, prompts, configuratiebestanden, package manifests, buildbestanden en installatiebestanden.

## Branch en releasebasis

1. Gebruik de huidige `dev` als bron voor de release.
2. Controleer eerst de status van `dev`.
3. Zorg dat de release exact gebaseerd is op de actuele `dev`.
4. `dev` moet de default branch blijven.
5. Wijzig de default branch niet naar een release-branch of tag.
6. Maak de release-tag:
   `V2.0.1`
7. Zorg dat de tag verwijst naar de correcte commit van `dev`.
8. Push de tag naar de remote als dat binnen de beschikbare rechten en repository-workflow is toegestaan.

## Versienummering

Controleer en corrigeer alle relevante versienummers zodat ze overeenkomen met `2.0.1`.

Controleer minimaal:

- applicatieversie
- package/project manifests
- Windows-buildconfiguratie
- `.exe`-builds/installers
- Flatpak metadata
- Snap metadata
- Docker image/tag configuratie
- CI/CD workflows
- release workflows
- installatie-instructies
- README/documentatie
- prompts
- changelog/release notes
- eventuele update- of version-checklogica

Er mogen geen oude releaseversies blijven staan waar die naar de huidige versie verwijzen.

## Installatiebestanden en distributie

De release moet daadwerkelijk installeerbaar zijn voor alle ondersteunde distributievormen.

### Windows

Controleer dat alle Windows releasebestanden correct vanuit de release beschikbaar zijn.

Minimaal:

- `.exe`
- eventuele Windows installer
- eventuele portable Windows-versie
- correcte bestandsnamen
- correcte versienummers
- correcte release-artifacts
- correcte links/verwijzingen in de documentatie

Controleer dat de Windows-bestanden daadwerkelijk bij de release worden gepubliceerd en niet alleen lokaal worden gebouwd.

### Flatpak

Voeg Flatpak-releaseondersteuning toe of corrigeer deze indien al aanwezig.

Controleer minimaal:

- Flatpak manifest
- applicatie-ID
- versie
- buildconfiguratie
- dependencies
- metadata
- iconen/assets indien vereist
- release artifact
- installatie-instructies
- verwijzingen vanuit de release/documentatie

Zorg dat een gebruiker de Flatpak-versie vanuit de release kan installeren.

### Snap

Voeg Snap-releaseondersteuning toe of corrigeer deze indien al aanwezig.

Controleer minimaal:

- `snapcraft.yaml`
- versie
- package metadata
- buildconfiguratie
- dependencies
- channels/tracks indien van toepassing
- release artifact
- installatie-instructies
- verwijzingen vanuit de release/documentatie

Zorg dat een gebruiker de Snap-versie vanuit de release kan installeren.

### Docker

Zorg dat de release ook als Docker-image beschikbaar is.

Controleer minimaal:

- `Dockerfile`
- buildconfiguratie
- image metadata
- versie
- image tags
- registry-configuratie
- release workflow
- installatie/run-instructies

De release moet minimaal een versiegebonden Docker-tag krijgen die overeenkomt met `2.0.1`.

Gebruik waar passend ook een stabiele release-tag, bijvoorbeeld:

- `2.0.1`
- `latest`

Gebruik `latest` alleen als dit overeenkomt met de bestaande projectconventie.

## CI/CD

Controleer alle CI/CD-workflows.

Zorg dat de releaseworkflow:

1. vanaf de correcte release-tag kan draaien;
2. de Windows artifacts bouwt;
3. de Flatpak artifacts bouwt/publiceert;
4. de Snap artifacts bouwt/publiceert;
5. de Docker image bouwt/publiceert;
6. correcte versienummers gebruikt;
7. artifacts daadwerkelijk aan de GitHub/GitLab/etc. release koppelt, afhankelijk van de gebruikte hosting;
8. geen artifact met een verkeerde of oude versie publiceert.

Los eventuele ontbrekende configuratie op.

## Release notes

Maak een duidelijk releaseverhaal voor `V2.0.1`.

Beschrijf daarin:

- wat er veranderd is;
- welke nieuwe functionaliteit is toegevoegd;
- welke bugs/problemen zijn opgelost;
- eventuele verbeteringen;
- wijzigingen aan installatie/distributie;
- dat Windows beschikbaar is;
- dat Flatpak beschikbaar is;
- dat Snap beschikbaar is;
- dat Docker beschikbaar is;
- eventuele belangrijke breaking changes of migratie-informatie.

Baseer de inhoud uitsluitend op daadwerkelijke wijzigingen in de repository. Verzin geen features of fixes.

Gebruik hiervoor bij voorkeur de verschillen tussen de vorige release/tag en `dev`.

## Documentatie

Werk alle relevante documentatie bij.

Controleer vooral:

- README
- installatiehandleidingen
- release-documentatie
- prompts
- deployment-documentatie
- Docker-documentatie
- Windows-installatie
- Flatpak-installatie
- Snap-installatie
- versievoorbeelden
- downloadlinks
- artifactnamen

Alle versies, commando's en links moeten overeenkomen met `V2.0.1`.

## Validatie vóór release

Voer vóór het maken/publiceren van de release een volledige controle uit.

Controleer minimaal:

- [ ] `dev` is de bron van de release.
- [ ] `dev` blijft de default branch.
- [ ] release-tag `V2.0.1` wordt aangemaakt.
- [ ] tag verwijst naar de juiste commit.
- [ ] alle relevante versienummers zijn `2.0.1` / `V2.0.1`.
- [ ] Windows `.exe`/installer is correct gebouwd.
- [ ] Windows artifact wordt aan de release gekoppeld.
- [ ] Flatpak is correct gebouwd.
- [ ] Flatpak is vanuit de release te installeren.
- [ ] Snap is correct gebouwd.
- [ ] Snap is vanuit de release te installeren.
- [ ] Docker image is correct gebouwd.
- [ ] Docker image heeft een `2.0.1` release-tag.
- [ ] release notes beschrijven de daadwerkelijke wijzigingen.
- [ ] documentatie bevat geen verkeerde versienummers.
- [ ] prompts bevatten geen verkeerde versienummers.
- [ ] installatie-instructies zijn correct.
- [ ] CI/CD bevat geen fouten of oude releaseverwijzingen.
- [ ] alle release-artifacts zijn aanwezig en correct benoemd.

## Belangrijke regel

Maak de release pas definitief wanneer de release-artifacts daadwerkelijk overeenkomen met `V2.0.1`.

Een Git-tag alleen is onvoldoende.

De release moet als geheel kloppen:

`dev` → `V2.0.1` tag → build → Windows + Flatpak + Snap + Docker → release artifacts → release notes → documentatie.

Rapporteer na uitvoering kort:

1. welke commit is gereleased;
2. welke tag is aangemaakt;
3. welke artifacts zijn gepubliceerd;
4. welke installatievormen beschikbaar zijn;
5. welke wijzigingen in V2.0.1 zitten;
6. eventuele resterende problemen of waarschuwingen.
