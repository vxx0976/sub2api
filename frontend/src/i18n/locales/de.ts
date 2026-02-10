export default {
  // Home Page
  home: {
    viewOnGithub: 'Auf GitHub anzeigen',
    viewDocs: 'Dokumentation anzeigen',
    docs: 'Dokumentation',
    switchToLight: 'Zum hellen Modus wechseln',
    switchToDark: 'Zum dunklen Modus wechseln',
    dashboard: 'Dashboard',
    login: 'Anmelden',
    register: 'Registrieren',
    getStarted: 'API-Schluessel holen',
    goToDashboard: 'Zum Dashboard gehen',
    seo: {
      title: '{siteName} - Guenstiger AI-Coding-Relay (Claude Code / Codex CLI / Gemini CLI)',
      description:
        'One API key for Claude Code, Codex CLI, and Gemini CLI. Transparent pricing, real-time usage, and safety limits for beginners.',
      keywords:
        'Sub2API, AI coding relay, Vibe Coding, Claude Code, Codex CLI, Gemini CLI, Cursor, OpenCode, API key, transparent pricing, usage, limits'
    },
    nav: {
      home: 'Startseite',
      pricing: 'Preise',
      docs: 'Dokumentation',
      status: 'Status'

    },
    hero: {
      badge: 'Guenstiger AI-Coding-Relay',
      title: 'Claude Code Relay-Dienst',
      subtitle: 'Claude API-Weiterleitung basierend auf dem offiziellen Anthropic-Protokoll',
      subtitleDesc: 'Einfache Einrichtung, transparente Preise, Echtzeitnutzungsverfolgung.',
      viewPricing: 'Preise',
      supportedToolsLabel: 'Funktioniert mit verschiedenen Entwicklungstools und Umgebungen',
      supportedToolsHint: 'Anthropic support may vary across different tools',
      platformsLabel: 'Plattformen',
      stats: {
        standard: 'Gaengige Tools unterstuetzt',
        copy: 'Kopieren & ausfuehren',
        transparent: 'Transparente Abrechnung',
        protect: 'Sicherheitslimits',
        stable: 'Stabil & zuverlaessig'

      },
      terminal: {
        title: 'Schnellstart',
        copy: 'Kopieren',
        hint: 'Ersetzen Sie sk-xxx durch den API-Schluessel aus Ihrem Dashboard. Pfade/Befehle koennen je nach Version variieren — Details finden Sie in der Dokumentation/im Dashboard.',
        compatibility: 'Folgt den offiziellen Anthropic API- und Umgebungsvariablen-Konventionen',
        claude: {
          line1: '# Claude Code konfigurieren',
          line2: '# Starten',
          success: '✅ Claude Code is ready'
        },
        codex: {
          line1: '# ~/.codex/config.toml (Beispiel)',
          line2: '# ~/.codex/auth.json',
          line3: '# Starten'

        },
        gemini: {
          line1: '# Gemini CLI konfigurieren',
          line2: '# Starten',
          success: '✅ Gemini CLI is ready'
        }
      }
    },
    steps: {
      oneTitle: 'Anmelden',
      oneDesc: 'Erstellen Sie ein Konto und generieren Sie einen API-Schluessel',
      twoTitle: 'Claude Code konfigurieren',
      twoDesc: 'Konfiguration kopieren und verbinden',
      threeTitle: 'Vibe Coding starten',
      threeDesc: 'Code schreiben, Fehler beheben, Tests hinzufuegen'

    },
    freeTrial: {
      badge: 'Neuer Benutzer Testversion',
      title: 'Testguthaben erhalten',
      description: 'Kontaktieren Sie den Kundenservice fuer Testguthaben'

    },
    quickstart: {
      title: 'Schnellstart',
      description: 'Ersetzen Sie BASE_URL und API-Schluessel und fuehren Sie dann aus:',
      copy: 'Beispiel kopieren',
      snippetTitle: 'Terminal',
      tip1: 'Wenn Sie ein SDK verwenden, setzen Sie einfach die Base URL/den Endpunkt auf Ihren Host.',
      tip2: 'Fangen Sie klein an: Versuchen Sie zuerst einen kurzen Prompt, um zu pruefen, ob alles funktioniert.'

    },
    why: {
      title: 'Warum Sub2API',
      subtitle: 'Einige Dinge, die Anfaengern helfen, schneller zu starten.'

    },
    tags: {
      subscriptionToApi: 'Abonnement zu API',
      stickySession: 'Persistente Sitzung',
      realtimeBilling: 'Echtzeitabrechnung'

    },
    features: {
      unifiedGateway: 'Claude Code natives Protokoll',
      unifiedGatewayDesc: 'Keine Aenderungen am Client-Verhalten noetig — kopieren Sie einfach die Umgebungsvariablen.',
      multiAccount: 'Multi-Knoten-Relay mit automatischem Failover',
      multiAccountDesc: 'Reduziert Anfragefehler, vermeidet Single Points of Failure.',
      balanceQuota: 'Nutzungstransparenz',
      balanceQuotaDesc: 'Token-genaue Aufzeichnungen mit Aufschluesselung nach Modell/Tag und uebersichtlicherem Guthaben.',
      rateLimit: 'Kostenschutz',
      rateLimitDesc: 'Ratenlimits und taegliche/woechentliche/monatliche Obergrenzen helfen, Ueberraschungen zu vermeiden.',
      concurrency: 'Spitzenlast-Resilienz',
      concurrencyDesc: 'Parallelitaetsschutz und Warteschlangen-Strategien fuer stabilere Spitzenzeiten.',
      adminConsole: 'Transparente Regeln',
      adminConsoleDesc: 'Modellpreise/-raten sind sichtbar mit nachvollziehbaren Abrechnungseintraegen.'

    },
    metrics: {
      title: 'Unsere Prinzipien',
      subtitle: 'Lassen Sie den Relay die Komplexitaet handhaben — Sie konzentrieren sich aufs Entwickeln.',
      items: {
        setup: 'Einrichtung',
        setupDesc: 'Konfiguration kopieren und in Minuten starten.',
        cost: 'Wert',
        costDesc: 'Waehlen Sie Modelle nach Bedarf, um Kosten zu optimieren.',
        transparent: 'Transparenz',
        transparentDesc: 'Klare Abrechnungsregeln und detaillierte Nutzungsaufzeichnungen.',
        coverage: 'Abdeckung',
        coverageDesc: 'Passt zu gaengigen CLI/IDE-Workflows.'

      }
    },
    pricing: {
      title: 'Waehlen Sie einen passenden Tarif',
      subtitle:
        'Multiple packages for different needs. No hidden fees — you can always view detailed usage.',
      recommended: 'Empfohlen',
      buy: 'Jetzt kaufen',
      note: 'Tarifkarten dienen der schnellen Einfuehrung. Tatsaechliches Guthaben, Limits und Abrechnungsregeln haengen von Ihrer Instanz ab.',
      unitPrice: 'Nur ¥{price}/Dollar',
      monthlyCredit: 'Monatliches Limit',
      customCredit: 'Benutzerdefiniertes Guthaben',
      noPlans: 'Keine Tarife verfuegbar',
      features: {
        monthlyLimit: 'Monatliches Limit ${limit}',
        validity: '{days} Tage Gueltigkeit',
        unitPrice: 'Nur ¥{price}/Dollar'

      }
    },
    providers: {
      title: 'Unterstuetzte Modelle & Plattformen',
      description: 'Claude ist das primaer unterstuetzte Modell; andere Modelle sind optional.',
      supported: 'Unterstuetzt',
      soon: 'Bald',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: 'Mehr'

    },
    faq: {
      title: 'Haeufig gestellte Fragen',
      subtitle: 'Haeufige Fragen fuer Erstbenutzer.',
      q0: 'Wird die OpenAI / ChatGPT API unterstuetzt?',
      a0: 'Dieser Dienst bietet keine OpenAI / ChatGPT API-Kompatibilitaet. Wir basieren auf dem offiziellen Anthropic-Protokoll, hauptsaechlich fuer Claude Code und verwandte Entwicklungsszenarien.',
      q1: 'Muss ich zuerst APIs lernen?',
      a1: 'Nein. Waehlen Sie Ihr Tool und kopieren Sie die Konfiguration. Lesen Sie die Dokumentation, wenn Sie mehr Optionen moechten.',
      q2: 'Wo bekomme ich einen API-Schluessel?',
      a2: 'Melden Sie sich an und gehen Sie zu Dashboard → API Keys, um einen zu erstellen/kopieren.',
      q3: 'Welche Tools/Clients werden unterstuetzt?',
      a3: 'Derzeit wird Claude Code unterstuetzt. Funktioniert mit Terminal, VS Code usw.',
      q4: 'Wie funktioniert die Preisgestaltung?',
      a4: 'Abrechnungsregeln und Nutzungsdetails werden im Dashboard angezeigt. Sie koennen nach Modell/Tag filtern.',
      q5: 'Was passiert, wenn eine Anfrage fehlschlaegt?',
      a5: 'Ueberpruefen Sie zuerst Host-URL, API-Schluessel und Netzwerkverbindung, dann schauen Sie ins Dashboard oder kontaktieren Sie den Administrator.',
      more: 'Dokumentation lesen'

    },
    referral: {
      badge: 'Empfehlungsbonus',
      title: 'Freunde einladen, Belohnungen teilen',
      description: 'Laden Sie Freunde ein, ihren ersten Kauf zu taetigen, beide erhalten 10% des Tarifpreises als Bonus',
      registerNow: 'Jetzt registrieren',
      goInvite: 'Einladen',
      feature1: 'Unbegrenzte Einladungen',
      feature2: 'Sofortige Gutschrift',
      feature3: 'Belohnungen skalieren mit dem Tarif'

    },
    cta: {
      title: 'Bereit es auszuprobieren?',
      description: 'Anmelden, API-Schluessel generieren, Tool-Konfiguration kopieren und mit dem Programmieren beginnen.'

    },
    footer: {
      allRightsReserved: 'Alle Rechte vorbehalten.'

    }
  },

  // Console Home Page
  consoleHome: {
    title: 'Startseite',
    description: 'Konsole',
    welcome: 'Willkommen zurueck',
    balance: 'Guthaben',
    gettingStarted: {
      title: 'Erste Schritte',
      badge: 'Empfohlen',
      description: 'Lernen Sie die Plattform in 3 Minuten kennen'

    },
    announcements: 'Ankuendigungen',
    promotions: 'Aktionen',
    quickLinks: 'Schnelllinks',
    announcement1: 'Begrenztes Angebot: Alle Tarife mit Bonusguthaben, bis zu $40 extra',
    announcement2: 'Auflade-Stufen-Upgrade: bis zu 2,2-facher Multiplikator',
    rechargeBonus: {
      title: 'Aufladebonus',
      description: 'Gestaffelter Multiplikator - je mehr Sie aufladen, desto hoeher der Satz',
      tier1: '¥10~199,99 = 2,0x',
      tier2: '¥200~499,99 = 2,1x',
      tier3: '¥500+ = 2,2x',
      action: 'Jetzt aufladen',
      perDollar: '/Dollar'

    },
    referralReward: {
      title: 'Empfehlungsbelohnung',
      description: 'Laden Sie Freunde ein, ihren ersten Kauf zu taetigen, beide erhalten 10% des Tarifpreises',
      planItem: '¥{price} Tarif = ${reward} pro Einladung',
      moreInfo: 'Kontaktieren Sie den Support fuer weitere Tarife',
      action: 'Freunde einladen'

    },
    contact: {
      title: 'Kontakt',
      wechat: 'WeChat',
      telegram: 'Telegram',
      scanQr: 'Scannen, um WeChat hinzuzufügen'
    },
    links: {
      subscriptions: 'Abonnements',
      plans: 'Tarife kaufen',
      apiKeys: 'API-Schluessel',
      referral: 'Freunde einladen'

    }
  },

  // Pricing Page
  pricing: {
    subscriptionPlans: 'Abonnementtarife',
    otherPlans: 'Weitere Optionen',
    perDollar: '/Dollar',
    plans: {
      starter: {
        name: 'Starter',
        credit: '$22 Guthaben',
        unitPrice: '¥0,45/Dollar',
        f1: '~50+ Programmieraufgaben',
        f2: '30 Tage Gueltigkeit',
        f3: 'Alle Modelle unterstuetzt',
        f4: 'Perfekt fuer gelegentliche Nutzung'

      },
      lite: {
        name: 'Lite',
        credit: '$45 Guthaben',
        unitPrice: '¥0,44/Dollar',
        f1: '~120+ Programmieraufgaben',
        f2: '30 Tage Gueltigkeit',
        f3: 'Alle Modelle unterstuetzt',
        f4: 'Fuer den taeglichen Entwicklungsbedarf'

      },
      standard: {
        name: 'Standard',
        credit: '$120 Guthaben',
        unitPrice: '¥0,42/Dollar',
        f1: '~300+ Programmieraufgaben',
        f2: '30 Tage Gueltigkeit',
        f3: 'Alle Modelle unterstuetzt',
        f4: 'Am beliebtesten, bestes Preis-Leistungs-Verhaeltnis'

      },
      pro: {
        name: 'Pro',
        credit: '$240 Guthaben',
        unitPrice: '¥0,42/Dollar',
        f1: '~600+ Programmieraufgaben',
        f2: '30 Tage Gueltigkeit',
        f3: 'Alle Modelle unterstuetzt',
        f4: 'Fuer intensive Nutzung'

      }
    },
    paygo: {
      name: 'Nutzungsbasiert',
      tagline: 'Flexible Nutzung',
      asLowAs: 'Ab nur',
      description: 'Kein Abonnement erforderlich, zahlen Sie nur fuer das, was Sie nutzen, Guthaben laeuft nie ab',
      feature1: 'Aufladen und nutzen, je mehr Sie aufladen, desto mehr erhalten Sie',
      feature2: 'Guthaben laeuft nie ab',
      feature3: 'Alle Modelle unterstuetzt',
      feature4: 'Perfekt fuer gelegentliche Nutzung',
      action: 'Jetzt aufladen'

    },
    custom: {
      name: 'Dedizierte Leitung',
      tagline: 'Enterprise-Loesung',
      price: 'Individuelle Preise',
      description: 'Dedizierter Account-Pool und exklusive Leitung fuer Unternehmenskunden',
      feature1: 'Dedizierter Account-Pool',
      feature2: 'Exklusive Hochgeschwindigkeitsleitung',
      feature3: 'Hoehere Parallelitaetsunterstuetzung',
      feature4: 'Dedizierter technischer Support',
      action: 'Kontaktieren Sie uns',
      wechatTitle: 'QR-Code scannen fuer WeChat',
      wechatDesc: 'Kontaktieren Sie uns fuer den dedizierten Leitungsdienst'

    }
  },

  // Status Page
  status: {
    title: 'Systemstatus',
    description: 'Systemdienststatus anzeigen',
    allOperational: 'Alle Systeme betriebsbereit',
    partialOutage: 'Teilweiser Systemausfall',
    majorOutage: 'Schwerwiegender Systemausfall',
    checking: 'Wird geprueft...',
    operational: 'Betriebsbereit',
    degraded: 'Eingeschraenkte Leistung',
    down: 'Dienststoerung',
    unknown: 'Unbekannt',
    uptime: 'Verfuegbarkeit',
    daysAgo: 'vor {days} Tagen',
    today: 'Heute',
    lastUpdated: 'Zuletzt aktualisiert',
    dataFrom: 'Statusdaten von'

  },

  // Setup Wizard
  setup: {
    title: 'Sub2API Einrichtung',
    description: 'Konfigurieren Sie Ihre Sub2API-Instanz',
    database: {
      title: 'Datenbankkonfiguration',
      description: 'Verbinden Sie sich mit Ihrer PostgreSQL-Datenbank',
      host: 'Host',
      port: 'Port',
      username: 'Benutzername',
      password: 'Passwort',
      databaseName: 'Datenbankname',
      sslMode: 'SSL-Modus',
      passwordPlaceholder: 'Passwort',
      ssl: {
        disable: 'Deaktivieren',
        require: 'Erforderlich',
        verifyCa: 'CA verifizieren',
        verifyFull: 'Vollstaendig verifizieren'

      }
    },
    redis: {
      title: 'Redis-Konfiguration',
      description: 'Verbinden Sie sich mit Ihrem Redis-Server',
      host: 'Host',
      port: 'Port',
      password: 'Passwort (optional)',
      database: 'Datenbank',
      passwordPlaceholder: 'Passwort',
      enableTls: 'TLS aktivieren',
      enableTlsHint: 'TLS bei Verbindung zu Redis verwenden (oeffentliche CA-Zertifikate)'

    },
    admin: {
      title: 'Admin-Konto',
      description: 'Erstellen Sie Ihr Administratorkonto',
      email: 'E-Mail',
      password: 'Passwort',
      confirmPassword: 'Passwort bestaetigen',
      passwordPlaceholder: 'Mindestens 6 Zeichen',
      confirmPasswordPlaceholder: 'Passwort bestaetigen',
      passwordMismatch: 'Passwoerter stimmen nicht ueberein'

    },
    ready: {
      title: 'Bereit zur Installation',
      description: 'Ueberpruefen Sie Ihre Konfiguration und schliessen Sie die Einrichtung ab',
      database: 'Datenbank',
      redis: 'Redis',
      adminEmail: 'Admin-E-Mail'

    },
    status: {
      testing: 'Wird getestet...',
      success: 'Verbindung erfolgreich',
      testConnection: 'Verbindung testen',
      installing: 'Wird installiert...',
      completeInstallation: 'Installation abschliessen',
      completed: 'Installation abgeschlossen!',
      redirecting: 'Weiterleitung zur Anmeldeseite...',
      restarting: 'Der Dienst wird neu gestartet, bitte warten...',
      timeout: 'Der Dienstneustart dauert laenger als erwartet. Bitte aktualisieren Sie die Seite manuell.'

    }
  },

  // Common
  common: {
    loading: 'Wird geladen...',
    save: 'Speichern',
    cancel: 'Abbrechen',
    delete: 'Loeschen',
    edit: 'Bearbeiten',
    create: 'Erstellen',
    update: 'Aktualisieren',
    confirm: 'Bestaetigen',
    reset: 'Zuruecksetzen',
    search: 'Suchen',
    filter: 'Filtern',
    export: 'Exportieren',
    import: 'Importieren',
    actions: 'Aktionen',
    status: 'Status',
    name: 'Name',
    email: 'E-Mail',
    password: 'Passwort',
    submit: 'Absenden',
    back: 'Zurueck',
    next: 'Weiter',
    yes: 'Ja',
    no: 'Nein',
    all: 'Alle',
    none: 'Keine',
    noData: 'Keine Daten',
    expand: 'Expand',
    collapse: 'Collapse',
    success: 'Erfolg',
    error: 'Fehler',
    critical: 'Kritisch',
    warning: 'Warnung',
    info: 'Info',
    active: 'Aktiv',
    inactive: 'Inaktiv',
    more: 'Mehr',
    close: 'Schliessen',
    enabled: 'Aktiviert',
    disabled: 'Deaktiviert',
    total: 'Gesamt',
    days: 'T',
    balance: 'Guthaben',
    available: 'Verfuegbar',
    copiedToClipboard: 'In Zwischenablage kopiert',
    copied: 'Kopiert',
    copyFailed: 'Kopieren fehlgeschlagen',
    verifying: 'Wird verifiziert...',
    processing: 'Wird verarbeitet...',
    contactSupport: 'Support kontaktieren',
    add: 'Hinzufuegen',
    invalidEmail: 'Bitte geben Sie eine gueltige E-Mail-Adresse ein',
    optional: 'optional',
    selectOption: 'Option auswaehlen',
    searchPlaceholder: 'Suchen...',
    noOptionsFound: 'Keine Optionen gefunden',
    noGroupsAvailable: 'Keine Gruppen verfuegbar',
    unknownError: 'Unbekannter Fehler aufgetreten',
    saving: 'Wird gespeichert...',
    selectedCount: '({count} ausgewaehlt)',
    refresh: 'Aktualisieren',
    settings: 'Einstellungen',
    notAvailable: 'N/A',
    now: 'Jetzt',
    unknown: 'Unbekannt',
    minutes: 'Min',
    reseller: 'Händler',
    time: {
      never: 'Nie',
      justNow: 'Gerade eben',
      minutesAgo: 'vor {n}min',
      hoursAgo: 'vor {n}h',
      daysAgo: 'vor {n}T',
      countdown: {
        daysHours: '{d}T {h}h',
        hoursMinutes: '{h}h {m}min',
        minutes: '{m}min',
        withSuffix: '{time} bis Aufhebung'

      }
    }
  },

  // Navigation
  nav: {
    consoleHome: 'Startseite',
    dashboard: 'Dashboard',
    announcements: 'Ankuendigungen',
    apiKeys: 'API-Schluessel',
    usage: 'Nutzung',
    redeem: 'Einloesen',
    profile: 'Profil',
    users: 'Benutzer',
    groups: 'Gruppen',
    subscriptions: 'Abonnements',
    accounts: 'Konten',
    proxies: 'Proxys',
    channels: 'Kanaele',
    redeemCodes: 'Einloesecodes',
    ops: 'Betrieb',
    promoCodes: 'Aktionscodes',
    settings: 'Einstellungen',
    myAccount: 'Mein Konto',
    lightMode: 'Heller Modus',
    darkMode: 'Dunkler Modus',
    collapse: 'Collapse',
    expand: 'Expand',
    logout: 'Abmelden',
    github: 'GitHub',
    mySubscriptions: 'Meine Abonnements',
    plans: 'Tarife',
    orders: 'Meine Bestellungen',
    adminOrders: 'Bestellverwaltung',
    adminReferrals: 'Empfehlungsverwaltung',
    adminRechargeOrders: 'Aufladeverwaltung',
    rechargeSettings: 'Aufladeeinstellungen',
    buySubscription: 'Abonnement kaufen',
    docs: 'Dokumentation',
    referral: 'Freunde einladen',
    resellerDashboard: 'Händler-Dashboard',
    resellerGroups: 'Pakete',
    resellerKeys: 'API-Schlüssel',
    resellerDomains: 'Domains',
    resellerSettings: 'Einstellungen',
    resellerSites: 'Website-Verwaltung'

  },

  // Docs
  docs: {
    title: 'Dokumentation',
    subtitle: 'Erfahren Sie, wie Sie den Dienst konfigurieren und nutzen',
    copy: 'Kopieren',
    copied: 'Kopiert',
    quickStart: {
      title: 'Schnellstart',
      step1: {
        title: 'API-Schluessel holen',
        desc: 'Erstellen Sie einen neuen API-Schluessel auf der "API-Schluessel"-Seite Ihres Dashboards'

      },
      step2: {
        title: 'Umgebung konfigurieren',
        desc: 'Richten Sie den API-Schluessel und die Base URL in Ihrem Tool ein'

      },
      step3: {
        title: 'Programmieren starten',
        desc: 'Starten Sie Ihr Tool und beginnen Sie Ihre AI-Coding-Reise'

      }
    },
    claudeCode: {
      title: 'Claude Code Konfiguration',
      desc: 'Setzen Sie die folgenden Umgebungsvariablen in Ihrem Terminal und starten Sie dann Claude Code:',
      replaceKey: 'Ersetzen Sie durch Ihren API-Schluessel'

    },
    models: {
      title: 'Unterstuetzte Modelle',
      opus: 'Leistungsstaerkstes Modell fuer komplexe Aufgaben',
      sonnet: 'Ausgewogene Leistung und Kosten',
      sonnet4: 'Kosteneffizienter Programmierassistent',
      haiku: 'Schnell und leichtgewichtig fuer einfache Aufgaben'

    },
    faq: {
      title: 'Haeufig gestellte Fragen',
      q1: {
        question: 'Wie wird das Guthaben berechnet?',
        answer: 'Guthaben wird basierend auf der tatsaechlichen Token-Nutzung abgerechnet. Verschiedene Modelle haben unterschiedliche Preismultiplikatoren. Spezifische Preise finden Sie auf der "Tarife"-Seite.'

      },
      q2: {
        question: 'Welche Tools werden unterstuetzt?',
        answer: 'Wir unterstuetzen Claude Code, Codex CLI, Gemini CLI und andere gaengige AI-Coding-Tools. Jedes Tool, das benutzerdefinierte API-Endpunkte unterstuetzt, kann verwendet werden.'

      },
      q3: {
        question: 'Was tun bei Problemen?',
        answer: 'Bitte ueberpruefen Sie zuerst, ob Ihr API-Schluessel und die Base URL korrekt konfiguriert sind. Wenn das Problem weiterhin besteht, kontaktieren Sie den Support.'

      }
    },
    contact: {
      title: 'Brauchen Sie Hilfe?',
      desc: 'Bei Fragen kontaktieren Sie bitte den Support oder erstellen Sie ein Ticket im Dashboard.'

    },
    guide: {
      title: 'Claude Code Einrichtungsanleitung',
      subtitle: 'Drei einfache Schritte zur Verbindung mit dem beschleunigten Netzwerk',
      note: 'Diese Anleitung ist fuer Claude Code (CLI)-Benutzer, basierend auf dem nativen Anthropic API-Protokoll.',
      warning: {
        title: 'Wichtig',
        intro: 'Diese API verwendet das native Anthropic-Protokoll:',
        item1: 'Funktioniert nur mit claude-code CLI und verwandten Entwicklungstools',
        item2: 'Unterstuetzt nicht die claude.ai-Weboberflaeche',
        item3: 'Unterstuetzt keine Tools, die auf der OpenAI API basieren (z.B. Cursor)'

      },
      step1: {
        title: 'Claude CLI installieren',
        description: 'Waehlen Sie die Installationsmethode fuer Ihr Betriebssystem:',
        commentBash: '# Claude CLI herunterladen und installieren',
        commentPS: '# Claude CLI herunterladen und installieren',
        commentCMD: ':: Claude CLI ueber npm installieren',
        tipMac: 'Moeglicherweise muessen Sie Ihr Terminal neu starten oder ausfuehren',
        tipPS: 'Moeglicherweise muessen Sie PowerShell nach der Installation neu starten',
        tipCMD: 'Stellen Sie sicher, dass Node.js installiert ist, und starten Sie CMD neu',
        verify: 'Installation verifizieren',
        verifySuccess: 'Wenn eine Versionsnummer angezeigt wird, war die Installation erfolgreich',
        networkTip: 'Wenn die Installation langsam ist, ueberpruefen Sie, ob Sie auf claude.ai zugreifen koennen'

      },
      step2: {
        title: 'API-Schluessel erstellen und konfigurieren',
        instruction1: 'Melden Sie sich an und gehen Sie zu',
        instruction1Link: 'Dashboard → API Keys',
        instruction2: 'Klicken Sie oben rechts auf "Schluessel erstellen", geben Sie einen Namen ein und bestaetigen Sie',
        instruction3Pre: 'Klicken Sie auf die',
        instruction3Button: 'Schluessel verwenden',
        instruction3Post: 'Schaltflaeche neben dem neuen Schluessel',
        instruction4: 'Waehlen Sie Ihr Betriebssystem und kopieren Sie die Umgebungsvariablen-Befehle',
        instruction5: 'Fuegen Sie die Befehle in Ihr Terminal ein und fuehren Sie sie aus',
        exampleTitle: '"Schluessel verwenden"-Dialog-Beispiel:',
        commentBash: '# Im Terminal kopieren und ausfuehren',
        commentPS: '# In PowerShell kopieren und ausfuehren',
        commentCMD: ':: In CMD kopieren und ausfuehren',
        yourKey: 'sk-Ihr-Schluessel',
        tipMac: 'Wir empfehlen, diese hinzuzufuegen zu',
        tipMacSuffix: 'fuer dauerhafte Konfiguration',
        tipMacOr: 'oder',
        tipPS: 'Wir empfehlen, diese zu Ihrem PowerShell-Profil hinzuzufuegen',
        tipPSSuffix: 'fuer dauerhafte Konfiguration',
        tipCMD: 'Wir empfehlen, dauerhafte Umgebungsvariablen ueber Systemeigenschaften → Umgebungsvariablen festzulegen',
        verify: 'Konfiguration verifizieren',
        verifyComment: '# Umgebungsvariablen ueberpruefen',
        verifyCMDComment: ':: Umgebungsvariablen ueberpruefen',
        verifyFail: 'Wenn keine Ausgabe erfolgt, sind die Umgebungsvariablen nicht gesetzt. Ueberpruefen Sie die Konfiguration oder starten Sie das Terminal neu.'

      },
      step3: {
        title: 'Claude starten',
        description: 'Fuehren Sie nach der Konfiguration den folgenden Befehl aus, um Claude zu starten:',
        tip: 'Der erste Start kann einige Sekunden zur Initialisierung dauern',
        debugTip: 'Wenn der Start fehlschlaegt, versuchen Sie',
        debugTipSuffix: 'um detaillierte Fehlerinformationen zu sehen'

      },
      tips: {
        title: 'Tipps',
        tip1: 'Umgebungsvariablen werden sofort wirksam, kein Terminalneustart erforderlich',
        tip2: 'Wir empfehlen, dauerhafte Umgebungsvariablen festzulegen, um Neukonfiguration zu vermeiden',
        tip3Pre: 'Probleme? Pruefen Sie die',
        tip3Link: 'Systemstatus',
        tip3Post: 'oder kontaktieren Sie den Support'

      },
      faq: {
        title: 'Haeufig gestellte Fragen',
        q1: 'F: 401 / Nicht autorisiert?',
        a1: 'Ueberpruefen Sie, ob ANTHROPIC_AUTH_TOKEN korrekt ist und keine zusaetzlichen Leerzeichen enthaelt.',
        q2: 'F: Verbindungstimeout / Netzwerkfehler?',
        a2: 'Ueberpruefen Sie, ob ANTHROPIC_BASE_URL korrekt ist und von Ihrem Netzwerk aus erreichbar ist.',
        q3: 'F: Claude startet, aber Gespraeche scheitern?',
        a3: 'Stellen Sie sicher, dass Sie die neueste Version von claude-code haben und nicht den Web-Login-Modus verwenden.'

      }
    },
    // Back button
    backToList: 'Zurueck zur Dokumentation',

    // Entry page
    entry: {
      title: 'Waehlen Sie Ihre Einrichtungsmethode',
      subtitle: 'Waehlen Sie die Konfigurationsanleitung fuer Ihr bevorzugtes Tool',
      recommended: 'Empfohlen',
      advanced: 'Erweitert',
      viewGuide: 'Anleitung anzeigen',
      vscodeDesc: 'Verwenden Sie Claude zum Programmieren in VSCode mit Plugins',
      cliDesc: 'Verwenden Sie das Claude Code Befehlszeilentool in Ihrem Terminal',
      tipsTitle: 'Tipps & Tricks',
      tipsDesc: 'Haeufige Befehle, Tastenkuerzel und Best Practices',
      quickInfo: {
        title: 'In 3 Schritten starten',
        step1Title: 'API-Schluessel holen',
        step1Desc: 'Erstellen Sie Ihren API-Schluessel im Dashboard',
        step2Title: 'Tool konfigurieren',
        step2Desc: 'Base URL und API-Schluessel einrichten',
        step3Title: 'Programmieren starten',
        step3Desc: 'Starten Sie Ihr Tool und geniessen Sie AI-Coding'

      },
      faq: {
        title: 'Haeufig gestellte Fragen',
        q1: 'Nicht sicher, welches Tool Sie waehlen sollen?',
        a1: 'Als Anfaenger empfehlen wir Cursor, da es eine grafische Oberflaeche bietet und leicht zu bedienen ist. Wenn Sie die Befehlszeile bevorzugen, probieren Sie Claude Code CLI.',
        q2: 'Ist es kostenpflichtig?',
        a2: 'Dieser Dienst verwendet nutzungsbasierte Preise mit unterschiedlichen Raten fuer verschiedene Modelle. Spezifische Preise finden Sie auf der "Tarife"-Seite.',
        q3: 'Welche Modelle werden unterstuetzt?',
        a3: 'Wir unterstuetzen Claude Opus 4.5, Sonnet 4.5 und andere gaengige Modelle. Sie koennen Modelle mit dem Befehl /model wechseln.',
        q4: 'Was tun bei Problemen?',
        a4: 'Ueberpruefen Sie zuerst, ob Ihr API-Schluessel und die Base URL korrekt konfiguriert sind. Wenn Probleme weiterhin bestehen, pruefen Sie die "Systemstatus"-Seite oder kontaktieren Sie den Support.'

      }
    },
    // VSCode setup guide
    vscode: {
      title: 'VSCode Einrichtungsanleitung',
      subtitle: 'Verwenden Sie Plugins, um Claude in Visual Studio Code zu integrieren',
      plugins: {
        title: 'Empfohlene Plugins',
        desc: 'Hier sind zwei beliebte Claude-Plugins, installieren Sie eines davon:',
        cline: 'Leistungsstarker AI-Coding-Assistent mit autonomem Programmieren, Dateioperationen usw.',
        claudeCode: 'Offizielles Anthropic VSCode-Plugin, konsistent mit der CLI-Erfahrung',
        official: 'Offiziell',
        install: 'Install'
      },
      step1: {
        title: 'Plugin installieren',
        instruction1: 'VSCode-Erweiterungen oeffnen (Strg+Umschalt+X)',
        instruction2: 'Suchen Sie nach "Cline" oder "Claude Code"',
        instruction3: 'Klicken Sie auf Installieren',
        tip: 'Beide Plugins sind aehnlich. Cline hat eine aktive Community und umfangreiche Funktionen, Claude Code ist offiziell mit einheitlicher Erfahrung'

      },
      step2: {
        title: 'API-Schluessel holen',
        instruction1: 'Melden Sie sich an und gehen Sie zu',
        instruction1Link: 'API-Schluessel-Verwaltung',
        instruction2: 'API-Schluessel erstellen und kopieren',
        infoTitle: 'Bitte beachten Sie Folgendes:',
        yourKey: 'sk-Ihr-Schluessel'

      },
      step3: {
        title: 'Plugin konfigurieren',
        selectPlugin: 'Waehlen Sie die Konfiguration fuer Ihr installiertes Plugin:',
        clineTitle: 'Cline-Konfiguration:',
        cline1: 'Klicken Sie auf das Cline-Symbol in der Seitenleiste, waehlen Sie "Eigenen API-Schluessel mitbringen" und klicken Sie auf Weiter',
        cline2: 'Konfigurieren Sie den API-Anbieter mit den folgenden Einstellungen und klicken Sie dann auf Weiter:',
        cline3: 'Konfiguration abgeschlossen! Sie koennen jetzt loslegen',
        claudeCodeTitle: 'Claude Code Konfiguration:',
        claudeCodeDesc: 'Das Claude Code Plugin verwendet Umgebungsvariablen, wie bei der CLI-Konfiguration:',
        claudeCode1: 'Setzen Sie die Umgebungsvariable ANTHROPIC_BASE_URL auf diese Siteadresse',
        claudeCode2: 'Setzen Sie die Umgebungsvariable ANTHROPIC_AUTH_TOKEN auf Ihren Schluessel',
        claudeCode3: 'Starten Sie VSCode neu, damit die Umgebungsvariablen wirksam werden',
        claudeCodeTipTitle: 'Tipp: ',
        claudeCodeTip: 'Weitere Informationen zur Umgebungsvariablenkonfiguration finden Sie in der CLI-Einrichtungsanleitung. Das VSCode-Plugin liest sie automatisch'

      },
      step4: {
        title: 'Konfiguration verifizieren',
        desc: 'Geben Sie eine Nachricht im Plugin-Chatfenster ein, um zu testen, ob es korrekt antwortet.',
        success: 'Wenn Sie eine Antwort erhalten, ist die Konfiguration erfolgreich!'

      },
      faq: {
        title: 'Haeufig gestellte Fragen',
        q1: 'Plugin kann keine Verbindung herstellen?',
        a1: 'Ueberpruefen Sie, ob das Base URL-Format korrekt ist (keine abschliessenden Schraegstriche), und ueberpruefen Sie, ob der API-Schluessel gueltig ist.',
        q2: 'Wie verwendet man verschiedene Modelle?',
        a2: 'Aendern Sie das Modellfeld in der Plugin-Konfiguration. Optionen sind u.a. claude-sonnet-4-20250514, claude-opus-4-20250514 usw.'

      }
    },
    // CLI setup guide
    cli: {
      title: 'Claude Code CLI Setup',
      subtitle: 'Verwenden Sie das offizielle Claude Code Befehlszeilentool in Ihrem Terminal',
      note: 'Diese Anleitung ist fuer claude-code CLI-Benutzer, basierend auf dem nativen Anthropic API-Protokoll.',
      prereq: {
        title: 'Voraussetzungen',
        windows: 'Windows-Benutzer:',
        windowsNode: 'Node.js muss zuerst installiert werden:',
        windowsTerminal: 'Wir empfehlen PowerShell oder Windows Terminal',
        mac: 'macOS / Linux-Benutzer:',
        macTerminal: 'Terminal oeffnen (Spotlight-Suche "Terminal")'

      }
    },
    // Tips page
    tips: {
      title: 'Tipps & Tricks',
      subtitle: 'Beherrschen Sie diese Befehle und Tipps fuer effizienteres AI-Coding',
      commands: {
        title: 'Befehlsreferenz',
        command: 'Befehl',
        function: 'Funktion',
        usage: 'Verwendungszweck',
        modelFunc: 'Modell wechseln',
        modelUsage: 'Verwenden Sie Opus fuer komplexe Aufgaben, Sonnet fuer den taeglichen Gebrauch',
        compactFunc: 'Kontext komprimieren',
        compactUsage: 'Wenn das Gespraech zu lang wird, spart Tokens',
        clearFunc: 'Gespraech loeschen',
        clearUsage: 'Beim Starten eines neuen Themas',
        costFunc: 'Kosten anzeigen',
        costUsage: 'Token-Nutzung fuer die aktuelle Sitzung pruefen',
        helpFunc: 'Hilfe anzeigen',
        helpUsage: 'Alle verfuegbaren Befehle anzeigen'

      },
      shortcuts: {
        title: 'Tastenkuerzel',
        shortcut: 'Tastenkuerzel',
        function: 'Funktion',
        escFunc: 'Aktuelle Operation abbrechen',
        shiftTabFunc: 'Planmodus beenden',
        ctrlCFunc: 'Erzwungenes Beenden'

      },
      models: {
        title: 'Modellauswahl-Tipps',
        desc: 'Waehlen Sie das richtige Modell basierend auf der Aufgabenkomplexitaet, um Leistung und Kosten auszugleichen:',
        sonnetLabel: 'Taegliche Wahl',
        sonnetDesc: 'Schnelle Antwort, kosteneffektiv, geeignet fuer die meisten Programmieraufgaben',
        sonnetUse1: 'Taegliches Code-Schreiben',
        sonnetUse2: 'Code-Review und Refactoring',
        sonnetUse3: 'Schnelle Fragen & Antworten und Debugging',
        opusLabel: 'Premium-Wahl',
        opusDesc: 'Staerkeres Reasoning, geeignet fuer komplexe Aufgaben',
        opusUse1: 'Komplexes Architekturdesign',
        opusUse2: 'Schwierige Problemloesung',
        opusUse3: 'Langtextanalyse',
        switchTip: 'Tipp:',
        switchDesc: 'Verwenden Sie den Befehl zum schnellen Modellwechsel'

      },
      practices: {
        title: 'Best Practices',
        tip1Title: 'Klaren Kontext liefern',
        tip1Desc: 'Erklaeren Sie vor der Frage Ihren Projekthintergrund, Tech-Stack und spezifische Anforderungen. Die AI gibt dann genauere Antworten.',
        tip2Title: '/compact klug einsetzen',
        tip2Desc: 'Wenn Gespraeche lang werden, verwenden Sie /compact zum Komprimieren des Verlaufs. Das spart Tokens und erhaelt die Antwortgeschwindigkeit.',
        tip3Title: 'Code-Snippets teilen',
        tip3Desc: 'Das Einfuegen relevanter Code-Snippets als Referenz fuer die AI ist effektiver als die reine Problembeschreibung.',
        tip4Title: 'Iterative Entwicklung',
        tip4Desc: 'Teilen Sie grosse Aufgaben in kleinere Schritte auf und arbeiten Sie schrittweise mit der AI zusammen. Das funktioniert besser als komplexe Features auf einmal anzufordern.'

      },
      links: {
        title: 'Schnelllinks',
        keys: 'API-Schluessel-Verwaltung',
        usage: 'Nutzung anzeigen',
        status: 'Systemstatus'

      }
    }
  },

  // Auth
  auth: {
    welcomeBack: 'Willkommen zurueck',
    signInToAccount: 'Melden Sie sich bei Ihrem Konto an, um fortzufahren',
    signIn: 'Anmelden',
    signingIn: 'Anmeldung...',
    createAccount: 'Konto erstellen',
    signUpToStart: 'Registrieren Sie sich, um {siteName} zu nutzen',
    signUp: 'Registrieren',
    processing: 'Wird verarbeitet...',
    continue: 'Continue',
    rememberMe: 'Angemeldet bleiben',
    dontHaveAccount: "Noch kein Konto?",
    alreadyHaveAccount: 'Bereits ein Konto?',
    registrationDisabled: 'Die Registrierung ist derzeit deaktiviert. Bitte kontaktieren Sie den Administrator.',
    emailLabel: 'E-Mail',
    emailPlaceholder: 'E-Mail-Adresse eingeben',
    passwordLabel: 'Passwort',
    passwordPlaceholder: 'Passwort eingeben',
    createPasswordPlaceholder: 'Erstellen Sie ein starkes Passwort',
    passwordHint: 'Mindestens 6 Zeichen',
    emailRequired: 'E-Mail ist erforderlich',
    invalidEmail: 'Bitte geben Sie eine gueltige E-Mail-Adresse ein',
    passwordRequired: 'Passwort ist erforderlich',
    passwordMinLength: 'Passwort muss mindestens 6 Zeichen lang sein',
    loginFailed: 'Anmeldung fehlgeschlagen. Bitte ueberpruefen Sie Ihre Zugangsdaten und versuchen Sie es erneut.',
    registrationFailed: 'Registrierung fehlgeschlagen. Bitte versuchen Sie es erneut.',
    loginSuccess: 'Anmeldung erfolgreich! Willkommen zurueck.',
    accountCreatedSuccess: 'Konto erfolgreich erstellt! Willkommen bei {siteName}.',
    reloginRequired: 'Sitzung abgelaufen. Bitte melden Sie sich erneut an.',
    turnstileExpired: 'Verifizierung abgelaufen, bitte versuchen Sie es erneut',
    turnstileFailed: 'Verifizierung fehlgeschlagen, bitte versuchen Sie es erneut',
    completeVerification: 'Bitte fuehren Sie die Verifizierung durch',
    verifyYourEmail: 'Verifizieren Sie Ihre E-Mail',
    sessionExpired: 'Sitzung abgelaufen',
    sessionExpiredDesc: 'Bitte gehen Sie zurueck zur Registrierungsseite und beginnen Sie erneut.',
    verificationCode: 'Verifizierungscode',
    verificationCodeHint: 'Geben Sie den 6-stelligen Code ein, der an Ihre E-Mail gesendet wurde',
    sendingCode: 'Wird gesendet...',
    clickToResend: 'Klicken, um den Code erneut zu senden',
    resendCode: 'Verifizierungscode erneut senden',
    promoCodeLabel: 'Aktionscode',
    promoCodePlaceholder: 'Aktionscode eingeben (optional)',
    promoCodeValid: 'Gueltig! Sie erhalten ${amount} Bonusguthaben',
    promoCodeInvalid: 'Ungueltiger Aktionscode',
    promoCodeNotFound: 'Aktionscode nicht gefunden',
    promoCodeExpired: 'Dieser Aktionscode ist abgelaufen',
    promoCodeDisabled: 'Dieser Aktionscode ist deaktiviert',
    promoCodeMaxUsed: 'Dieser Aktionscode hat sein Nutzungslimit erreicht',
    promoCodeAlreadyUsed: 'Sie haben diesen Aktionscode bereits verwendet',
    promoCodeValidating: 'Aktionscode wird ueberprueft, bitte warten',
    promoCodeInvalidCannotRegister: 'Ungueltiger Aktionscode. Bitte ueberpruefen und erneut versuchen oder das Aktionscode-Feld leeren',
    invitationCodeLabel: 'Einladungscode',
    invitationCodePlaceholder: 'Einladungscode eingeben',
    invitationCodeRequired: 'Einladungscode ist erforderlich',
    invitationCodeValid: 'Einladungscode ist gueltig',
    invitationCodeInvalid: 'Ungueltiger oder bereits verwendeter Einladungscode',
    invitationCodeValidating: 'Einladungscode wird ueberprueft...',
    invitationCodeInvalidCannotRegister: 'Ungueltiger Einladungscode. Bitte ueberpruefen und erneut versuchen',
    linuxdo: {
      signIn: 'Weiter mit Linux.do',
      orContinue: 'oder weiter mit E-Mail',
      callbackTitle: 'Anmeldung laeuft',
      callbackProcessing: 'Anmeldung wird abgeschlossen, bitte warten...',
      callbackHint: 'Wenn Sie nicht automatisch weitergeleitet werden, gehen Sie zurueck zur Anmeldeseite und versuchen Sie es erneut.',
      callbackMissingToken: 'Fehlender Anmelde-Token, bitte versuchen Sie es erneut.',
      backToLogin: 'Zurueck zur Anmeldung'

    },
    oauth: {
      code: 'Code',
      state: 'Status',
      fullUrl: 'Vollstaendige URL'

    },
    // Forgot password
    forgotPassword: 'Passwort vergessen?',
    forgotPasswordTitle: 'Passwort zuruecksetzen',
    forgotPasswordHint: 'Geben Sie Ihre E-Mail-Adresse ein und wir senden Ihnen einen Link zum Zuruecksetzen Ihres Passworts.',
    sendResetLink: 'Link zum Zuruecksetzen senden',
    sendingResetLink: 'Wird gesendet...',
    sendResetLinkFailed: 'Link zum Zuruecksetzen konnte nicht gesendet werden. Bitte versuchen Sie es erneut.',
    resetEmailSent: 'Link zum Zuruecksetzen gesendet',
    resetEmailSentHint: 'Wenn ein Konto mit dieser E-Mail existiert, erhalten Sie in Kuerze einen Link zum Zuruecksetzen des Passworts. Bitte pruefen Sie Ihren Posteingang und Spam-Ordner.',
    backToLogin: 'Zurueck zur Anmeldung',
    rememberedPassword: 'Passwort wieder eingefallen?',

    // Reset password
    resetPasswordTitle: 'Neues Passwort festlegen',
    resetPasswordHint: 'Geben Sie unten Ihr neues Passwort ein.',
    newPassword: 'Neues Passwort',
    newPasswordPlaceholder: 'Neues Passwort eingeben',
    confirmPassword: 'Passwort bestaetigen',
    confirmPasswordPlaceholder: 'Confirm your new password',
    confirmPasswordRequired: 'Bitte bestaetigen Sie Ihr Passwort',
    passwordsDoNotMatch: 'Passwoerter stimmen nicht ueberein',
    resetPassword: 'Passwort zuruecksetzen',
    resettingPassword: 'Resetting...',
    resetPasswordFailed: 'Failed to reset password. Please try again.',
    passwordResetSuccess: 'Password Reset Successful',
    passwordResetSuccessHint: 'Ihr Passwort wurde zurueckgesetzt. Sie koennen sich jetzt mit Ihrem neuen Passwort anmelden.',
    invalidResetLink: 'Invalid Reset Link',
    invalidResetLinkHint: 'Dieser Link zum Zuruecksetzen des Passworts ist ungueltig oder abgelaufen. Bitte fordern Sie einen neuen an.',
    requestNewResetLink: 'Request New Reset Link',
    invalidOrExpiredToken: 'Der Link zum Zuruecksetzen des Passworts ist ungueltig oder abgelaufen. Bitte fordern Sie einen neuen an.',

    // Email verification page
    sendCodeTo: "Wir senden einen Verifizierungscode an",
    codeSentSuccess: 'Verifizierungscode gesendet! Bitte pruefen Sie Ihren Posteingang.',
    verifying: 'Wird verifiziert...',
    verifyAndCreate: 'Verify & Create Account',
    resendCodeIn: 'Code erneut senden in {seconds}s',
    backToRegistration: 'Back to registration',
    sendCodeFailed: 'Failed to send verification code. Please try again.',
    pleaseCompleteVerification: 'Bitte fuehren Sie die Verifizierung durch',
    codeRequired: 'Verifizierungscode ist erforderlich',
    codeInvalid: 'Bitte geben Sie einen gueltigen 6-stelligen Code ein',
    verificationFailed: 'Verification failed. Please try again.'
  },

  // Dashboard
  dashboard: {
    title: 'Dashboard',
    welcomeMessage: "Willkommen zurueck! Hier ist eine Uebersicht Ihres Kontos.",
    balance: 'Guthaben',
    apiKeys: 'API-Schluessel',
    todayRequests: 'Today Requests',
    todayCost: 'Today Cost',
    todayTokens: 'Today Tokens',
    totalTokens: 'Gesamt-Tokens',
    cacheToday: 'Cache (Today)',
    performance: 'Performance',
    avgResponse: 'Avg Response',
    averageTime: 'Average time',
    timeRange: 'Time Range',
    granularity: 'Granularity',
    day: 'Tag',
    hour: 'Stunde',
    modelDistribution: 'Model Distribution',
    tokenUsageTrend: 'Token Usage Trend',
    noDataAvailable: 'No data available',
    model: 'Modell',
    requests: 'Anfragen',
    tokens: 'Tokens',
    actual: 'Actual',
    standard: 'Standard',
    input: 'Eingabe',
    output: 'Ausgabe',
    cache: 'Cache',
    recentUsage: 'Recent Usage',
    last7Days: 'Letzte 7 Tage',
    noUsageRecords: 'No usage records',
    startUsingApi: 'Beginnen Sie mit der API-Nutzung, um hier Ihren Nutzungsverlauf zu sehen.',
    viewAllUsage: 'View all usage',
    quickActions: 'Quick Actions',
    createApiKey: 'Create API Key',
    generateNewKey: 'Generate a new API key',
    viewUsage: 'Nutzung anzeigen',
    checkDetailedLogs: 'Detaillierte Nutzungsprotokolle anzeigen',
    redeemCode: 'Einloesecode',
    addBalanceWithCode: 'Guthaben mit einem Code hinzufuegen'

  },

  // Groups (shared)
  groups: {
    subscription: 'Sub'

  },

  // API Keys
  keys: {
    title: 'API-Schluessel',
    description: 'Manage your API keys and access tokens',
    createKey: 'Create API Key',
    editKey: 'Edit API Key',
    deleteKey: 'Delete API Key',
    deleteConfirmMessage: "Moechten Sie '{name}' wirklich loeschen? Diese Aktion kann nicht rueckgaengig gemacht werden.",
    apiKey: 'API-Schluessel',
    group: 'Gruppe',
    noGroup: 'No group',
    created: 'Erstellt',
    copyToClipboard: 'Copy to clipboard',
    copied: 'Kopiert!',
    importToCcSwitch: 'Import to CCS',
    enable: 'Aktivieren',
    disable: 'Deaktivieren',
    nameLabel: 'Name',
    namePlaceholder: 'My API Key',
    notes: 'Notizen',
    notesPlaceholder: 'Optionale Notizen',
    groupLabel: 'Gruppe',
    selectGroup: 'Select a group',
    statusLabel: 'Status',
    selectStatus: 'Select status',
    saving: 'Wird gespeichert...',
    noKeysYet: 'No API keys yet',
    createFirstKey: 'Create your first API key to get started with the API.',
    keyCreatedSuccess: 'API-Schluessel erfolgreich erstellt',
    keyUpdatedSuccess: 'API-Schluessel erfolgreich aktualisiert',
    keyDeletedSuccess: 'API-Schluessel erfolgreich geloescht',
    keyEnabledSuccess: 'API-Schluessel erfolgreich aktiviert',
    keyDisabledSuccess: 'API-Schluessel erfolgreich deaktiviert',
    failedToLoad: 'Failed to load API keys',
    failedToSave: 'Failed to save API key',
    failedToDelete: 'Failed to delete API key',
    failedToUpdateStatus: 'Failed to update API key status',
    clickToChangeGroup: 'Klicken, um die Gruppe zu aendern',
    groupChangedSuccess: 'Gruppe erfolgreich geaendert',
    failedToChangeGroup: 'Failed to change group',
    groupRequired: 'Bitte waehlen Sie eine Gruppe',
    usage: 'Nutzung',
    today: 'Heute',
    total: 'Gesamt',
    quota: 'Kontingent',
    useKey: 'Schluessel verwenden',
    useKeyModal: {
      title: 'Use API Key',
      description:
        'Add the following environment variables to your terminal profile or run directly in terminal to configure API access.',
      copy: 'Kopieren',
      copied: 'Kopiert',
      note: 'These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.',
      noGroupTitle: 'Bitte weisen Sie zuerst eine Gruppe zu',
      noGroupDescription: 'Diesem API-Schluessel wurde keine Gruppe zugewiesen. Bitte klicken Sie auf die Gruppenspalte in der Schluesselliste, um eine zuzuweisen, bevor Sie die Konfiguration anzeigen.',
      openai: {
        description: 'Fuegen Sie die folgenden Konfigurationsdateien zu Ihrem Codex CLI-Konfigurationsverzeichnis hinzu.',
        configTomlHint: 'Stellen Sie sicher, dass der folgende Inhalt am Anfang der config.toml-Datei steht',
        note: 'Stellen Sie sicher, dass das Konfigurationsverzeichnis existiert. macOS/Linux-Benutzer koennen mkdir -p ~/.codex ausfuehren.',
        noteWindows: 'Press Win+R and enter %userprofile%\\.codex to open the config directory. Create it manually if it does not exist.',
      },
      cliTabs: {
        claudeCode: 'Claude Code',
        geminiCli: 'Gemini CLI',
        codexCli: 'Codex CLI',
        opencode: 'OpenCode',
      },
      antigravity: {
        description: 'Konfigurieren Sie den API-Zugang fuer die Antigravity-Gruppe. Waehlen Sie die Konfigurationsmethode basierend auf Ihrem Client.',
        claudeCode: 'Claude Code',
        geminiCli: 'Gemini CLI',
        claudeNote: 'These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.',
        geminiNote: 'These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.',
      },
      gemini: {
        description: 'Fuegen Sie die folgenden Umgebungsvariablen zu Ihrem Terminal-Profil hinzu oder fuehren Sie sie direkt im Terminal aus, um den Gemini CLI-Zugang zu konfigurieren.',
        modelComment: 'Wenn Sie Zugang zu Gemini 3 haben, koennen Sie gemini-3-pro-preview verwenden',
        note: 'These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.',
      },
      opencode: {
        title: 'OpenCode-Beispiel',
        subtitle: 'opencode.json',
        hint: 'Config path: ~/.config/opencode/opencode.json (or opencode.jsonc), create if not exists. Use default providers (openai/anthropic/google) or custom provider_id. API Key can be configured directly or via /connect command. This is an example, adjust models and options as needed.',
      },
    },
    customKeyLabel: 'Benutzerdefinierter Schluessel',
    customKeyPlaceholder: 'Geben Sie Ihren benutzerdefinierten Schluessel ein (min. 16 Zeichen)',
    customKeyHint: 'Nur Buchstaben, Zahlen, Unterstriche und Bindestriche erlaubt. Mindestens 16 Zeichen.',
    customKeyTooShort: 'Benutzerdefinierter Schluessel muss mindestens 16 Zeichen lang sein',
    customKeyInvalidChars: 'Benutzerdefinierter Schluessel darf nur Buchstaben, Zahlen, Unterstriche und Bindestriche enthalten',
    customKeyRequired: 'Bitte geben Sie einen benutzerdefinierten Schluessel ein',
    ipRestriction: 'IP Restriction',
    ipWhitelist: 'IP-Whitelist',
    ipWhitelistPlaceholder: '192.168.1.100\n10.0.0.0/8',
    ipWhitelistHint: 'Eine IP oder CIDR pro Zeile. Nur diese IPs koennen diesen Schluessel verwenden.',
    ipBlacklist: 'IP-Blacklist',
    ipBlacklistPlaceholder: '1.2.3.4\n5.6.0.0/16',
    ipBlacklistHint: 'Eine IP oder CIDR pro Zeile. Diese IPs werden von der Verwendung dieses Schluessels ausgeschlossen.',
    ipRestrictionEnabled: 'IP-Beschraenkung aktiviert',
    ccSwitchNotInstalled: 'CC-Switch ist nicht installiert oder der Protokoll-Handler ist nicht registriert. Bitte installieren Sie zuerst CC-Switch oder kopieren Sie den API-Schluessel manuell.',
    ccsClientSelect: {
      title: 'Select Client',
      description: 'Bitte waehlen Sie den Client-Typ fuer den Import in CC-Switch:',
      claudeCode: 'Claude Code',
      claudeCodeDesc: 'Import as Claude Code configuration',
      geminiCli: 'Gemini CLI',
      geminiCliDesc: 'Import as Gemini CLI configuration',
    },
    // Quota and expiration
    quotaLimit: 'Quota Limit',
    quotaAmount: 'Quota Amount (USD)',
    quotaAmountPlaceholder: 'Kontingentlimit in USD eingeben',
    quotaAmountHint: 'Legen Sie den Hoechstbetrag fest, den dieser Schluessel ausgeben kann. 0 = unbegrenzt.',
    quotaUsed: 'Quota Used',
    reset: 'Zuruecksetzen',
    resetQuotaUsed: 'Reset used quota to 0',
    resetQuotaTitle: 'Confirm Reset Quota',
    resetQuotaConfirmMessage: 'Moechten Sie das verbrauchte Kontingent (${used}) fuer den Schluessel "{name}" wirklich auf 0 zuruecksetzen? Diese Aktion kann nicht rueckgaengig gemacht werden.',
    quotaResetSuccess: 'Quota reset successfully',
    failedToResetQuota: 'Failed to reset quota',
    expiration: 'Ablauf',
    expiresInDays: '{days} Tage',
    extendDays: '+{days} Tage',
    customDate: 'Benutzerdefiniert',
    expirationDate: 'Expiration Date',
    expirationDateHint: 'Waehlen Sie, wann dieser API-Schluessel ablaufen soll.',
    currentExpiration: 'Aktuelles Ablaufdatum',
    expiresAt: 'Laeuft ab',
    noExpiration: 'Nie',
    status: {
      active: 'Aktiv',
      inactive: 'Inaktiv',
      quota_exhausted: 'Quota Exhausted',
      expired: 'Abgelaufen',

    },
  },

  // Usage
  usage: {
    title: 'Usage Records',
    description: 'API-Nutzungsverlauf anzeigen und analysieren',
    costDetails: 'Cost Breakdown',
    tokenDetails: 'Token Breakdown',
    totalRequests: 'Total Requests',
    totalTokens: 'Gesamt-Tokens',
    totalCost: 'Total Cost',
    standardCost: 'Standard',
    actualCost: 'Actual',
    userBilled: 'User billed',
    accountBilled: 'Account billed',
    accountMultiplier: 'Account rate',
    avgDuration: 'Avg Duration',
    inSelectedRange: 'in selected range',
    perRequest: 'per request',
    apiKeyFilter: 'API-Schluessel',
    allApiKeys: 'All API Keys',
    timeRange: 'Time Range',
    exportCsv: 'Export CSV',
    exportExcel: 'Export Excel',
    exportingProgress: 'Exporting data...',
    exportedCount: 'Exported {current}/{total} records',
    estimatedTime: 'Geschaetzte verbleibende Zeit: {time}',
    cancelExport: 'Cancel Export',
    exportCancelled: 'Export cancelled',
    exporting: 'Exporting...',
    preparingExport: 'Preparing export...',
    model: 'Modell',
    reasoningEffort: 'Reasoning Effort',
    type: 'Typ',
    tokens: 'Tokens',
    cost: 'Kosten',
    firstToken: 'First Token',
    duration: 'Duration',
    time: 'Zeit',
    stream: 'Stream',
    sync: 'Sync',
    in: 'In',
    out: 'Aus',
    cacheRead: 'Gelesen',
    cacheWrite: 'Schreiben',
    rate: 'Rate',
    original: 'Original',
    billed: 'Billed',
    noRecords: 'No usage records found. Try adjusting your filters.',
    failedToLoad: 'Failed to load usage logs',
    noDataToExport: 'No data to export',
    exportSuccess: 'Usage data exported successfully',
    exportFailed: 'Failed to export usage data',
    exportExcelSuccess: 'Usage data exported successfully (Excel format)',
    exportExcelFailed: 'Failed to export usage data',
    imageUnit: ' Bilder',
    userAgent: 'User-Agent'

  },

  // Redeem
  redeem: {
    title: 'Einloesecode',
    description: 'Geben Sie Ihren Einloesecode ein, um Guthaben hinzuzufuegen oder die Parallelitaet zu erhoehen',
    currentBalance: 'Current Balance',
    concurrency: 'Parallelitaet',
    requests: 'requests',
    redeemCodeLabel: 'Einloesecode',
    redeemCodePlaceholder: 'Geben Sie Ihren Einloesecode ein',
    redeemCodeHint: 'Redeem codes are case-sensitive',
    redeeming: 'Redeeming...',
    redeemButton: 'Einloesecode',
    redeemSuccess: 'Code Redeemed Successfully!',
    redeemFailed: 'Redemption Failed',
    added: 'Hinzugefuegt',
    concurrentRequests: 'concurrent requests',
    newBalance: 'New Balance',
    newConcurrency: 'New Concurrency',
    aboutCodes: 'About Redeem Codes',
    codeRule1: 'Jeder Code kann nur einmal verwendet werden',
    codeRule2: 'Codes koennen Guthaben hinzufuegen, Parallelitaet erhoehen oder Testzugang gewaehren',
    codeRule3: 'Kontaktieren Sie den Support bei Problemen beim Einloesen eines Codes',
    codeRule4: 'Balance and concurrency updates are immediate',
    recentActivity: 'Recent Activity',
    historyWillAppear: 'Ihr Einloeseverlauf wird hier angezeigt',
    balanceAddedRedeem: 'Balance Added (Redeem)',
    balanceAddedAdmin: 'Balance Added (Admin)',
    balanceDeductedAdmin: 'Balance Deducted (Admin)',
    concurrencyAddedRedeem: 'Concurrency Added (Redeem)',
    concurrencyAddedAdmin: 'Concurrency Added (Admin)',
    concurrencyReducedAdmin: 'Concurrency Reduced (Admin)',
    adminAdjustment: 'Admin-Anpassung',
    subscriptionAssigned: 'Subscription Assigned',
    subscriptionAssignedDesc: 'Ihnen wurde Zugang zu {groupName} gewaehrt',
    subscriptionDays: '{days} Tage',
    days: ' Tage',
    codeRedeemSuccess: 'Code erfolgreich eingeloest!',
    failedToRedeem: 'Failed to redeem code. Please check the code and try again.',
    subscriptionRefreshFailed: 'Redeemed successfully, but failed to refresh subscription status.',
    pleaseEnterCode: 'Bitte geben Sie einen Einloesecode ein'

  },

  // Referral
  referral: {
    title: 'Freunde einladen',
    subtitle: 'Invite friends to join and both get rewards',
    yourCode: 'Your Referral Code',
    shareCodeHint: 'Teilen Sie diesen Code mit Freunden, sie koennen ihn bei der Registrierung verwenden',
    totalInvited: 'Total Invited',
    totalRewarded: 'Rewarded',
    pendingPayment: 'Pending Payment',
    totalEarnings: 'Total Earnings',
    rewardRules: 'Reward Rules',
    rule1: 'Freunde geben Ihren Empfehlungscode bei der Registrierung ein (im Aktionscode-Feld)',
    rule2: 'Wenn Freunde ihren ersten Kauf taetigen, erhalten beide Seiten 10% des Tarifpreises als Bonus',
    rule3: 'Belohnungen werden sofort Ihrem Kontoguthaben gutgeschrieben',
    rule4: 'Belohnungen werden nur bei der ersten Zahlung ausgestellt; Aufladungen und Einloesecodes zaehlen nicht',
    inviteeList: 'Invitation Records',
    noInvitees: 'Noch keine Einladungseintraege',
    inviteeEmail: 'E-Mail',
    joinedAt: 'Joined At',
    status: 'Status',
    yourEarning: 'Your Earning',
    statusRewarded: 'Rewarded',
    statusPending: 'Ausstehend',
    showingPage: 'Page {current} of {total}',
    loadCodeFailed: 'Failed to load referral code',
    linkCopied: 'Empfehlungslink kopiert',
    referralCodeValid: 'Empfehlungscode gueltig',
    referralCodeHint: 'Erhalten Sie 10% des Tarifpreises als Bonus nach Ihrem ersten Kauf',
    referralCodeInvalid: 'Empfehlungscode ist ungueltig oder abgelaufen',
    referralCodeSelf: 'Sie koennen Ihren eigenen Empfehlungscode nicht verwenden'

  },

  // Profile
  profile: {
    title: 'Profile Settings',
    description: 'Manage your account information and settings',
    accountBalance: 'Account Balance',
    concurrencyLimit: 'Concurrency Limit',
    memberSince: 'Member Since',
    administrator: 'Administrator',
    user: 'Benutzer',
    username: 'Benutzername',
    enterUsername: 'Benutzername eingeben',
    editProfile: 'Edit Profile',
    updateProfile: 'Update Profile',
    updating: 'Wird aktualisiert...',
    updateSuccess: 'Profile updated successfully',
    updateFailed: 'Failed to update profile',
    usernameRequired: 'Username is required',
    changePassword: 'Passwort aendern',
    currentPassword: 'Aktuelles Passwort',
    newPassword: 'Neues Passwort',
    confirmNewPassword: 'Confirm New Password',
    passwordHint: 'Password must be at least 8 characters long',
    changingPassword: 'Wird geaendert...',
    changePasswordButton: 'Passwort aendern',
    passwordsNotMatch: 'Neue Passwoerter stimmen nicht ueberein',
    passwordTooShort: 'Password must be at least 8 characters long',
    passwordChangeSuccess: 'Password changed successfully',
    passwordChangeFailed: 'Failed to change password',
    // TOTP 2FA
    totp: {
      title: 'Zwei-Faktor-Authentifizierung (2FA)',
      description: 'Kontosicherheit mit Google Authenticator oder aehnlichen Apps erhoehen',
      enabled: 'Aktiviert',
      enabledAt: 'Enabled at',
      notEnabled: 'Not Enabled',
      notEnabledHint: 'Enable two-factor authentication to enhance account security',
      enable: 'Aktivieren',
      disable: 'Deaktivieren',
      featureDisabled: 'Funktion nicht verfuegbar',
      featureDisabledHint: 'Die Zwei-Faktor-Authentifizierung wurde vom Administrator nicht aktiviert',
      setupTitle: 'Zwei-Faktor-Authentifizierung einrichten',
      setupStep1: 'Scannen Sie den QR-Code unten mit Ihrer Authenticator-App',
      setupStep2: 'Geben Sie den 6-stelligen Code aus Ihrer App ein',
      manualEntry: "Can't scan? Enter the key manually:",
      enterCode: '6-stelligen Code eingeben',
      verify: 'Verifizieren',
      setupFailed: 'Failed to get setup information',
      verifyFailed: 'Ungueltiger Code, bitte versuchen Sie es erneut',
      enableSuccess: 'Zwei-Faktor-Authentifizierung aktiviert',
      disableTitle: 'Disable Two-Factor Authentication',
      disableWarning: 'Nach der Deaktivierung benoetigen Sie keinen Verifizierungscode mehr zum Anmelden. Dies kann Ihre Kontosicherheit verringern.',
      enterPassword: 'Enter your current password to confirm',
      confirmDisable: 'Confirm Disable',
      disableSuccess: 'Zwei-Faktor-Authentifizierung deaktiviert',
      disableFailed: 'Failed to disable, please check your password',
      loginTitle: 'Zwei-Faktor-Authentifizierung',
      loginHint: 'Geben Sie den 6-stelligen Code aus Ihrer Authenticator-App ein',
      loginFailed: 'Verifizierung fehlgeschlagen, bitte versuchen Sie es erneut',

      // New translations for email verification
      verifyEmailFirst: 'Bitte verifizieren Sie zuerst Ihre E-Mail',
      verifyPasswordFirst: 'Bitte verifizieren Sie zuerst Ihre Identitaet',
      emailCode: 'Email Verification Code',
      enterEmailCode: '6-stelligen Code eingeben',
      sendCode: 'Code senden',
      codeSent: 'Verifizierungscode an Ihre E-Mail gesendet',
      sendCodeFailed: 'Failed to send verification code'
    }
  },

  // Empty States
  empty: {
    noData: 'No data found'
  },

  // Table
  table: {
    expandActions: 'Expand More Actions',
    collapseActions: 'Collapse Actions'
  },

  // Pagination
  pagination: {
    showing: 'Showing',
    to: 'bis',
    of: 'von',
    results: 'results',
    page: 'Seite',
    pageOf: 'Page {page} of {total}',
    previous: 'Zurueck',
    next: 'Weiter',
    perPage: 'Per page',
    goToPage: 'Gehe zu Seite {page}',
    jumpTo: 'Jump to',
    jumpPlaceholder: 'Seite',
    jumpAction: 'Go'
  },

  // Errors
  errors: {
    somethingWentWrong: 'Something went wrong',
    pageNotFound: 'Page not found',
    unauthorized: 'Nicht autorisiert',
    forbidden: 'Verboten',
    serverError: 'Server error',
    networkError: 'Network error',
    timeout: 'Request timeout',
    tryAgain: 'Please try again'
  },

  // Dates
  dates: {
    today: 'Heute',
    yesterday: 'Gestern',
    thisWeek: 'Diese Woche',
    lastWeek: 'Last Week',
    thisMonth: 'Dieser Monat',
    lastMonth: 'Last Month',
    last7Days: 'Letzte 7 Tage',
    last14Days: 'Last 14 Days',
    last30Days: 'Letzte 30 Tage',
    custom: 'Benutzerdefiniert',
    startDate: 'Startdatum',
    endDate: 'Enddatum',
    apply: 'Anwenden',
    selectDateRange: 'Select date range'
  },

  // Admin
  admin: {
    // Dashboard
    dashboard: {
      title: 'Admin Dashboard',
      description: 'Systemuebersicht und Echtzeitstatistiken',
      apiKeys: 'API-Schluessel',
      totalApiKeys: 'Total API Keys',
      activeApiKeys: 'Active Keys',
      users: 'Benutzer',
      totalUsers: 'Total Users',
      activeUsers: 'active users',
      accounts: 'Konten',
      totalAccounts: 'Total Accounts',
      activeAccounts: 'Active Accounts',
      todayRequests: 'Today Requests',
      totalRequests: 'Total Requests',
      todayCost: 'Today Cost',
      totalCost: 'Total Cost',
      actual: 'Actual',
      standard: 'Standard',
      todayTokens: 'Today Tokens',
      totalTokens: 'Gesamt-Tokens',
      input: 'Eingabe',
      output: 'Ausgabe',
      cacheToday: 'Cache (Today)',
      performance: 'Performance',
      avgResponse: 'Avg Response',
      averageTime: 'Avg Time',
      timeRange: 'Time Range',
      granularity: 'Granularity',
      day: 'Tag',
      hour: 'Stunde',
      modelDistribution: 'Model Distribution',
      tokenUsageTrend: 'Token Usage Trend',
      userUsageTrend: 'User Usage Trend (Top 12)',
      noDataAvailable: 'No data available',
      model: 'Modell',
      requests: 'Anfragen',
      tokens: 'Tokens',
      cache: 'Cache',
      recentUsage: 'Recent Usage',
      last7Days: 'Letzte 7 Tage',
      noUsageRecords: 'No usage records',
      startUsingApi: 'Usage history will appear here once you start using the API.',
      viewAllUsage: 'Alle anzeigen',
      quickActions: 'Quick Actions',
      manageUsers: 'Manage Users',
      viewUserAccounts: 'Benutzerkonten anzeigen und verwalten',
      manageAccounts: 'Manage Accounts',
      configureAiAccounts: 'KI-Plattform-Konten konfigurieren',
      systemSettings: 'System Settings',
      configureSystem: 'Systemeinstellungen konfigurieren',
      newUsersToday: 'New Users Today',
      active: 'active',
      ok: 'ok',
      err: 'Fehler',
      create: 'Erstellen',
      failedToLoad: 'Failed to load dashboard statistics'
    },

    // Users
    users: {
      title: 'User Management',
      description: 'Manage users and their permissions',
      createUser: 'Benutzer erstellen',
      editUser: 'Benutzer bearbeiten',
      deleteUser: 'Benutzer loeschen',
      searchUsers: 'Search users...',
      allRoles: 'Alle Rollen',
      allStatus: 'Alle Status',
      admin: 'Admin',
      user: 'Benutzer',
      disabled: 'Deaktiviert',
      email: 'E-Mail',
      password: 'Passwort',
      username: 'Benutzername',
      notes: 'Notizen',
      enterEmail: 'E-Mail eingeben',
      enterPassword: 'Passwort eingeben',
      enterUsername: 'Benutzername eingeben (optional)',
      enterNotes: 'Notizen eingeben (nur Admin)',
      notesHint: 'Diese Notiz ist nur fuer Administratoren sichtbar',
      enterNewPassword: 'Neues Passwort eingeben (optional)',
      leaveEmptyToKeep: 'Leave empty to keep current password',
      generatePassword: 'Generate random password',
      copyPassword: 'Copy password',
      creating: 'Wird erstellt...',
      updating: 'Wird aktualisiert...',
      columns: {
        user: 'Benutzer',
        email: 'E-Mail',
        username: 'Benutzername',
        notes: 'Notizen',
        role: 'Rolle',
        subscriptions: 'Abonnements',
        balance: 'Guthaben',
        usage: 'Nutzung',
        concurrency: 'Parallelitaet',
        status: 'Status',
        created: 'Erstellt',
        actions: 'Aktionen'

      },
      today: 'Heute',
      total: 'Gesamt',
      noSubscription: 'No subscription',
      daysRemaining: '{days}T',
      expired: 'Abgelaufen',
      disable: 'Deaktivieren',
      enable: 'Aktivieren',
      disableUser: 'Disable User',
      enableUser: 'Enable User',
      viewApiKeys: 'API-Schluessel anzeigen',
      groups: 'Gruppen',
      apiKeys: 'API-Schluessel',
      userApiKeys: 'User API Keys',
      noApiKeys: 'Dieser Benutzer hat keine API-Schluessel',
      group: 'Gruppe',
      none: 'Keine',
      noUsersYet: 'No users yet',
      createFirstUser: 'Create your first user to get started.',
      userCreated: 'User created successfully',
      userUpdated: 'User updated successfully',
      userDeleted: 'User deleted successfully',
      userEnabled: 'User enabled successfully',
      userDisabled: 'User disabled successfully',
      failedToLoad: 'Failed to load users',
      failedToCreate: 'Failed to create user',
      failedToUpdate: 'Failed to update user',
      failedToDelete: 'Failed to delete user',
      failedToToggle: 'Failed to update user status',
      failedToLoadApiKeys: 'Failed to load user API keys',
      emailRequired: 'Bitte E-Mail eingeben',
      concurrencyMin: 'Concurrency must be at least 1',
      amountRequired: 'Bitte geben Sie einen gueltigen Betrag ein',
      insufficientBalance: 'Unzureichendes Guthaben',
      deleteConfirm: "Moechten Sie '{email}' wirklich loeschen? Diese Aktion kann nicht rueckgaengig gemacht werden.",
      setAllowedGroups: 'Set Allowed Groups',
      allowedGroupsHint:
        'Select which standard groups this user can use. Subscription groups are managed separately.',
      noStandardGroups: 'Keine Standardgruppen verfuegbar',
      allowAllGroups: 'Allow All Groups',
      allowAllGroupsHint: 'User can use any non-exclusive group',
      allowedGroupsUpdated: 'Erlaubte Gruppen erfolgreich aktualisiert',
      failedToLoadGroups: 'Failed to load groups',
      failedToUpdateAllowedGroups: 'Failed to update allowed groups',
      // User Group Configuration
      groupConfig: 'User Group Configuration',
      groupConfigHint: 'Benutzerdefinierte Ratenmultiplikatoren fuer Benutzer {email} konfigurieren (ueberschreibt Gruppenstandards)',
      exclusiveGroups: 'Exclusive Groups',
      publicGroups: 'Public Groups (Default Available)',
      defaultRate: 'Default Rate',
      customRate: 'Custom Rate',
      useDefaultRate: 'Standard verwenden',
      customRatePlaceholder: 'Leer lassen fuer Standard',
      groupConfigUpdated: 'Gruppenkonfiguration erfolgreich aktualisiert',
      deposit: 'Einzahlung',
      withdraw: 'Auszahlung',
      depositAmount: 'Deposit Amount',
      withdrawAmount: 'Withdraw Amount',
      withdrawAll: 'Alle',
      currentBalance: 'Current Balance',
      depositNotesPlaceholder:
        'e.g., New user registration bonus, promotional credit, compensation, etc.',
      withdrawNotesPlaceholder:
        'e.g., Service issue refund, incorrect charge reversal, account closure refund, etc.',
      notesOptional: 'Notizen sind optional, aber hilfreich fuer die Buchfuehrung',
      amountHint: 'Bitte geben Sie einen positiven Betrag ein',
      newBalance: 'New Balance',
      depositing: 'Depositing...',
      withdrawing: 'Withdrawing...',
      confirmDeposit: 'Confirm Deposit',
      confirmWithdraw: 'Confirm Withdraw',
      depositSuccess: 'Deposit successful',
      withdrawSuccess: 'Withdraw successful',
      failedToDeposit: 'Failed to deposit',
      failedToWithdraw: 'Failed to withdraw',
      useDepositWithdrawButtons: 'Bitte verwenden Sie Einzahlen/Abheben-Schaltflaechen zur Guthaben-Anpassung',
      deleteConfirmMessage: "Moechten Sie den Benutzer '{email}' wirklich loeschen? Diese Aktion kann nicht rueckgaengig gemacht werden.",
      searchPlaceholder: 'Search users...',
      roleFilter: 'Role Filter',
      statusFilter: 'Status Filter',
      allStatuses: 'Alle Status',
      form: {
        emailLabel: 'E-Mail',
        emailPlaceholder: 'E-Mail eingeben',
        usernameLabel: 'Benutzername',
        usernamePlaceholder: 'Benutzername eingeben (optional)',
        notesLabel: 'Notizen',
        notesPlaceholder: 'Notizen eingeben (nur Admin)',
        notesHint: 'Diese Notiz ist nur fuer Administratoren sichtbar',
        passwordLabel: 'Passwort',
        passwordPlaceholder: 'Passwort eingeben (leer lassen zum Beibehalten)',
        roleLabel: 'Rolle',
        selectRole: 'Select role',
        balanceLabel: 'Guthaben',
        concurrencyLabel: 'Parallelitaet',
        statusLabel: 'Status',
        selectStatus: 'Select status'
      },
      adjustBalance: 'Adjust Balance',
      adjustConcurrency: 'Adjust Concurrency',
      adjustmentAmount: 'Anpassungsbetrag',
      adjustmentAmountHint: 'Positiv zum Hinzufuegen, negativ zum Abziehen',
      currentConcurrency: 'Current Concurrency',
      saving: 'Wird gespeichert...',
      noUsers: 'No users',
      noUsersDescription: 'Create your first user to get started.',
      userCreatedSuccess: 'User created successfully',
      userUpdatedSuccess: 'User updated successfully',
      userDeletedSuccess: 'User deleted successfully',
      balanceAdjustedSuccess: 'Balance adjusted successfully',
      concurrencyAdjustedSuccess: 'Concurrency adjusted successfully',
      failedToSave: 'Failed to save user',
      failedToAdjust: 'Failed to adjust',
      // Balance History
      balanceHistory: 'Aufladeverlauf',
      balanceHistoryTip: 'Klicken, um den Aufladeverlauf zu oeffnen',
      balanceHistoryTitle: 'User Recharge & Concurrency History',
      noBalanceHistory: 'Keine Eintraege fuer diesen Benutzer gefunden',
      allTypes: 'All Types',
      typeBalance: 'Balance (Redeem)',
      typeAdminBalance: 'Balance (Admin)',
      typeConcurrency: 'Concurrency (Redeem)',
      typeAdminConcurrency: 'Concurrency (Admin)',
      typeSubscription: 'Abonnement',
      failedToLoadBalanceHistory: 'Failed to load balance history',
      createdAt: 'Erstellt',
      totalRecharged: 'Total Recharged',
      roles: {
        admin: 'Admin',
        user: 'Benutzer',
        reseller: 'Händler'

      },
      // Settings Dropdowns
      filterSettings: 'Filter Settings',
      columnSettings: 'Column Settings',
      filterValue: 'Wert eingeben',

      // User Attributes
      attributes: {
        title: 'User Attributes',
        description: 'Benutzerdefinierte Benutzerattribut-Felder konfigurieren',
        configButton: 'Attribute',
        addAttribute: 'Attribut hinzufuegen',
        editAttribute: 'Edit Attribute',
        deleteAttribute: 'Delete Attribute',
        deleteConfirm: "Moechten Sie das Attribut '{name}' wirklich loeschen? Alle Benutzerwerte fuer dieses Attribut werden geloescht.",
        noAttributes: 'Keine benutzerdefinierten Attribute',
        noAttributesHint: 'Klicken Sie auf die Schaltflaeche oben, um benutzerdefinierte Attribute hinzuzufuegen',
        key: 'Attributschluessel',
        keyHint: 'Fuer programmatische Referenz, nur Buchstaben, Zahlen und Unterstriche',
        name: 'Anzeigename',
        nameHint: 'In Formularen angezeigter Name',
        type: 'Attribute Type',
        fieldDescription: 'Beschreibung',
        fieldDescriptionHint: 'Beschreibungstext fuer das Attribut',
        placeholder: 'Platzhalter',
        placeholderHint: 'Platzhaltertext fuer das Eingabefeld',
        required: 'Erforderlich',
        enabled: 'Aktiviert',
        options: 'Optionen',
        optionsHint: 'Fuer Auswahl-/Mehrfachauswahl-Typen',
        addOption: 'Option hinzufuegen',
        optionValue: 'Option Value',
        optionLabel: 'Anzeigetext',
        validation: 'Validierungsregeln',
        minLength: 'Min. Laenge',
        maxLength: 'Max. Laenge',
        min: 'Min Value',
        max: 'Max Value',
        pattern: 'Regex-Muster',
        patternMessage: 'Validation Error Message',
        types: {
          text: 'Text',
          textarea: 'Textbereich',
          number: 'Zahl',
          email: 'E-Mail',
          url: 'URL',
          date: 'Datum',
          select: 'Auswahl',
          multi_select: 'Mehrfachauswahl'

        },
        created: 'Attribut erfolgreich erstellt',
        updated: 'Attribut erfolgreich aktualisiert',
        deleted: 'Attribut erfolgreich geloescht',
        reordered: 'Attributreihenfolge erfolgreich aktualisiert',
        failedToLoad: 'Failed to load attributes',
        failedToCreate: 'Failed to create attribute',
        failedToUpdate: 'Failed to update attribute',
        keyRequired: 'Bitte Attributschluessel eingeben',
        nameRequired: 'Bitte Anzeigename eingeben',
        optionsRequired: 'Bitte mindestens eine Option hinzufuegen',
        failedToDelete: 'Failed to delete attribute',
        failedToReorder: 'Failed to update order',
        keyExists: 'Attribute key already exists',
        dragToReorder: 'Ziehen zum Sortieren'

      }
    },

    // Groups
    groups: {
      title: 'Group Management',
      description: 'Manage API key groups and rate multipliers',
      searchGroups: 'Search groups...',
      createGroup: 'Gruppe erstellen',
      editGroup: 'Gruppe bearbeiten',
      deleteGroup: 'Gruppe loeschen',
      allPlatforms: 'All Platforms',
      activePlans: 'Purchasable + Balance',
      allPurchasable: 'Alle',
      purchasable: 'Purchasable',
      notPurchasable: 'Not Purchasable',
      allStatus: 'Alle Status',
      allGroups: 'All Groups',
      exclusive: 'Exklusiv',
      nonExclusive: 'Non-Exclusive',
      public: 'Public',
      columns: {
        name: 'Name',
        platform: 'Plattform',
        rateMultiplier: 'Ratenmultiplikator',
        exclusive: 'Exklusiv',
        type: 'Typ',
        purchasable: 'Purchasable',
        priority: 'Prioritaet',
        apiKeys: 'API-Schluessel',
        accounts: 'Konten',
        status: 'Status',
        actions: 'Aktionen',
        billingType: 'Abrechnungstyp'

      },
      rateAndAccounts: '{rate}x Rate · {count} Konten',
      accountsCount: '{count} Konten',
      noAccounts: 'No accounts',
      form: {
        name: 'Name',
        description: 'Beschreibung',
        platform: 'Plattform',
        rateMultiplier: 'Ratenmultiplikator',
        status: 'Status',
        exclusive: 'Exclusive Group',
        nameLabel: 'Gruppenname',
        namePlaceholder: 'Enter group name',
        descriptionLabel: 'Beschreibung',
        descriptionPlaceholder: 'Enter description (optional)',
        rateMultiplierLabel: 'Ratenmultiplikator',
        rateMultiplierHint: '1,0 = Standardrate, 0,5 = halber Preis, 2,0 = doppelt',
        exclusiveLabel: 'Exclusive Group',
        exclusiveHint: 'Exclusive group, manually assign to users',
        platformLabel: 'Plattform-Beschraenkung',
        platformPlaceholder: 'Select platform (leave empty for no restriction)',
        accountsLabel: 'Assigned Accounts',
        accountsPlaceholder: 'Konten auswaehlen (leer lassen fuer keine Beschraenkung)',
        priorityLabel: 'Prioritaet',
        priorityHint: 'Niedrigerer Wert = hoehere Prioritaet, verwendet fuer Kontoplanung',
        statusLabel: 'Status'

      },
      exclusiveObj: {
        yes: 'Ja',
        no: 'Nein'

      },
      saving: 'Wird gespeichert...',
      noGroups: 'No groups',
      noGroupsDescription: 'Create groups to better manage API keys and rates.',
      groupCreatedSuccess: 'Gruppe erfolgreich erstellt',
      groupUpdatedSuccess: 'Gruppe erfolgreich aktualisiert',
      groupDeletedSuccess: 'Gruppe erfolgreich geloescht',
      failedToSave: 'Failed to save group',
      exclusiveFilter: 'Exklusiv',
      enterGroupName: 'Enter group name',
      optionalDescription: 'Optional description',
      platformHint: 'Waehlen Sie die Plattform, der diese Gruppe zugeordnet ist',
      platformNotEditable: 'Plattform kann nach der Erstellung nicht geaendert werden',
      rateMultiplierHint: 'Cost multiplier for this group (e.g., 1.5 = 150% of base cost)',
      exclusiveHint: 'Exclusive group, manually assign to specific users',
      exclusiveTooltip: {
        title: 'Was ist eine exklusive Gruppe?',
        description: 'Wenn aktiviert, koennen Benutzer diese Gruppe beim Erstellen von API-Schluesseln nicht sehen. Erst nachdem ein Admin einen Benutzer manuell dieser Gruppe zuweist, kann er sie nutzen.',
        example: 'Anwendungsfall:',
        exampleContent: 'Public group rate is 0.8. Create an exclusive group with 0.7 rate, manually assign VIP users to give them better pricing.'
      },
      noGroupsYet: 'No groups yet',
      createFirstGroup: 'Create your first group to organize API keys.',
      creating: 'Wird erstellt...',
      updating: 'Wird aktualisiert...',
      limitDay: 'T',
      limitWeek: 'W',
      limitMonth: 'Mo',
      groupCreated: 'Gruppe erfolgreich erstellt',
      groupUpdated: 'Gruppe erfolgreich aktualisiert',
      groupDeleted: 'Gruppe erfolgreich geloescht',
      failedToLoad: 'Failed to load groups',
      failedToCreate: 'Failed to create group',
      failedToUpdate: 'Failed to update group',
      failedToDelete: 'Failed to delete group',
      sortOrder: 'Sortierung per Drag & Drop',
      sortOrderHint: 'Gruppen ziehen, um die Anzeigereihenfolge anzupassen',
      sortOrderUpdated: 'Sortierung aktualisiert',
      failedToUpdateSortOrder: 'Sortierung konnte nicht aktualisiert werden',
      nameRequired: 'Bitte Gruppenname eingeben',
      platforms: {
        all: 'All Platforms',
        anthropic: 'Anthropic',
        openai: 'OpenAI',
        gemini: 'Gemini',
        antigravity: 'Antigravity'

      },
      deleteConfirm:
        "Are you sure you want to delete '{name}'? All associated API keys will no longer belong to any group.",
      deleteConfirmSubscription:
        "Are you sure you want to delete subscription group '{name}'? This will invalidate all API keys bound to this subscription and delete all related subscription records. This action cannot be undone.",
      subscription: {
        title: 'Subscription Settings',
        type: 'Abrechnungstyp',
        typeHint:
          'Standard billing deducts from user balance. Subscription mode uses quota limits instead.',
        typeNotEditable: 'Abrechnungstyp kann nach der Gruppenerstellung nicht geaendert werden.',
        standard: 'Standard (Balance)',
        subscription: 'Subscription (Quota)',
        dailyLimit: 'Daily Limit (USD)',
        weeklyLimit: 'Weekly Limit (USD)',
        monthlyLimit: 'Monthly Limit (USD)',
        defaultValidityDays: 'Standard-Gueltigkeit (Tage)',
        validityHint: 'Anzahl der Tage, die das Abonnement gueltig ist, wenn es einem Benutzer zugewiesen wird',
        noLimit: 'Kein Limit'

      },
      imagePricing: {
        title: 'Image Generation Pricing',
        description: 'Preise fuer das Modell gemini-3-pro-image konfigurieren. Leer lassen fuer Standardpreise.'

      },
      claudeCode: {
        title: 'Claude Code Client-Beschraenkung',
        tooltip: 'When enabled, this group only allows official Claude Code clients. Non-Claude Code requests will be rejected or fallback to the specified group.',
        enabled: 'Nur Claude Code',
        disabled: 'Alle Clients erlauben',
        fallbackGroup: 'Fallback-Gruppe',
        fallbackHint: 'Non-Claude Code requests will use this group. Leave empty to reject directly.',
        noFallback: 'No Fallback (Reject)'
      },
      invalidRequestFallback: {
        title: 'Fallback-Gruppe fuer ungueltige Anfragen',
        hint: 'Wird nur ausgeloest, wenn Upstream explizit meldet, dass der Prompt zu lang ist. Leer lassen zum Deaktivieren.',
        noFallback: 'No Fallback'
      },
      copyAccounts: {
        title: 'Copy Accounts from Groups',
        tooltip: 'Waehlen Sie eine oder mehrere Gruppen derselben Plattform. Nach der Erstellung werden alle Konten dieser Gruppen automatisch an die neue Gruppe gebunden (dedupliziert).',
        tooltipEdit: 'Select one or more groups of the same platform. After saving, current group accounts will be replaced with accounts from these groups (deduplicated).',
        selectPlaceholder: 'Gruppen auswaehlen, von denen Konten kopiert werden sollen...',
        hint: 'Mehrere Gruppen koennen ausgewaehlt werden, Konten werden dedupliziert',
        hintEdit: '⚠️ Warning: This will replace all existing account bindings'
      },
      modelRouting: {
        title: 'Modell-Routing',
        tooltip: 'Configure specific model requests to be routed to designated accounts. Supports wildcard matching, e.g., claude-opus-* matches all opus models.',
        enabled: 'Aktiviert',
        disabled: 'Deaktiviert',
        disabledHint: 'Routing-Regeln werden nur wirksam, wenn sie aktiviert sind',
        addRule: 'Routing-Regel hinzufuegen',
        modelPattern: 'Model Pattern',
        modelPatternPlaceholder: 'claude-opus-*',
        modelPatternHint: 'Unterstuetzt * Platzhalter, z.B. claude-opus-* passt auf alle Opus-Modelle',
        accounts: 'Priority Accounts',
        selectAccounts: 'Konten auswaehlen',
        noAccounts: 'No accounts in this group',
        loadingAccounts: 'Konten werden geladen...',
        removeRule: 'Remove Rule',
        noRules: 'Keine Routing-Regeln',
        noRulesHint: 'Add routing rules to route specific model requests to designated accounts',
        searchAccountPlaceholder: 'Search accounts...',
        accountsHint: 'Konten auswaehlen, die fuer dieses Modellmuster priorisiert werden sollen'

      },
      payment: {
        title: 'Plan Sales Settings',
        description: 'Gruppe als kaufbaren Tarif konfigurieren',
        validityDays: 'Gueltigkeit (Tage)',
        validityDaysHint: 'Anzahl der Tage, die das Abonnement nach dem Kauf gueltig ist',
        price: 'Preis (CNY)',
        priceHint: 'Der Betrag, den Benutzer fuer diesen Tarif zahlen',
        isPurchasable: 'Allow Purchase',
        isPurchasableHint: 'Wenn aktiviert, koennen Benutzer diese Gruppe auf der Tarifseite kaufen',
        sortOrder: 'Sortiergewicht',
        sortOrderHint: 'Hoehere Werte erscheinen zuerst in der Liste',
        isRecommended: 'Empfohlen',
        isRecommendedHint: 'Empfohlen-Badge auf der Tarifseite anzeigen',
        resellerTemplate: 'Händler-Vorlage',
        resellerTemplateHint: 'Wenn aktiviert, können Händler diesen Plan als Vorlage für ihre Benutzer verwenden',
        externalBuyUrl: 'Externer Kauf-Link',
        externalBuyUrlPlaceholder: 'https://item.taobao.com/item.htm?id=...',
        externalBuyUrlHint: 'Optional. Wenn konfiguriert, koennen Benutzer auf externen Plattformen kaufen'

      },
      planPreview: {
        title: 'Tarifvorschau',
        button: 'Preview Plans',
        noPlans: 'Keine Tarife anzuzeigen',
        noPlansDesc: 'Bitte erstellen Sie zuerst Abonnement-Gruppen und aktivieren Sie den Kauf.',
        notPurchasable: 'Not Purchasable',
        previewOnly: 'Preview Only'
      },
      mcpXml: {
        title: 'MCP XML-Protokoll-Injektion',
        tooltip: 'When enabled, if the request contains MCP tools, an XML format call protocol prompt will be injected into the system prompt. Disable this to avoid interference with certain clients.',
        enabled: 'Aktiviert',
        disabled: 'Deaktiviert'

      },
      supportedScopes: {
        title: 'Supported Model Families',
        tooltip: 'Waehlen Sie die Modellfamilien, die diese Gruppe unterstuetzt. Nicht ausgewaehlte Familien werden nicht an diese Gruppe weitergeleitet.',
        claude: 'Claude',
        geminiText: 'Gemini Text',
        geminiImage: 'Gemini Bild',
        hint: 'Waehlen Sie mindestens eine Modellfamilie'

      }
    },

    // Subscriptions
    subscriptions: {
      title: 'Subscription Management',
      description: 'Manage user subscriptions and quota limits',
      assignSubscription: 'Assign Subscription',
      extendSubscription: 'Extend Subscription',
      revokeSubscription: 'Revoke Subscription',
      allStatus: 'Alle Status',
      allGroups: 'All Groups',
      daily: 'Daily',
      weekly: 'Weekly',
      monthly: 'Monthly',
      noLimits: 'Keine Limits konfiguriert',
      unlimited: 'Unbegrenzt',
      resetNow: 'Resetting soon',
      windowNotActive: 'Window not active',
      resetInMinutes: 'Resets in {minutes}m',
      resetInHoursMinutes: 'Resets in {hours}h {minutes}m',
      resetInDaysHours: 'Resets in {days}d {hours}h',
      daysRemaining: 'days remaining',
      noExpiration: 'No expiration',
      status: {
        active: 'Aktiv',
        expired: 'Abgelaufen',
        revoked: 'Widerrufen'

      },
      columns: {
        user: 'Benutzer',
        group: 'Gruppe',
        source: 'Quelle',
        usage: 'Nutzung',
        expires: 'Laeuft ab',
        status: 'Status',
        actions: 'Aktionen'

      },
      source: {
        admin: 'Admin Assigned',
        purchase: 'Kauf',
        redeem: 'Einloesecode',
        unknown: 'Unbekannt'

      },
      sourceHistory: 'Source History',
      form: {
        user: 'Benutzer',
        group: 'Abonnement-Gruppe',
        validityDays: 'Gueltigkeit (Tage)',
        extendDays: 'Extend by (Days)'
      },
      selectUser: 'Benutzer auswaehlen',
      selectGroup: 'Abonnement-Gruppe auswaehlen',
      groupHint: 'Nur Gruppen mit Abonnement-Abrechnungstyp werden angezeigt',
      validityHint: 'Anzahl der Tage, die das Abonnement gueltig sein wird',
      extendingFor: 'Extending subscription for',
      currentExpiration: 'Aktuelles Ablaufdatum',
      assign: 'Zuweisen',
      assigning: 'Assigning...',
      extend: 'Verlaengern',
      extending: 'Extending...',
      revoke: 'Widerrufen',
      noSubscriptionsYet: 'No subscriptions yet',
      assignFirstSubscription: 'Assign a subscription to get started.',
      subscriptionAssigned: 'Abonnement erfolgreich zugewiesen',
      subscriptionExtended: 'Abonnement erfolgreich verlaengert',
      subscriptionRevoked: 'Abonnement erfolgreich widerrufen',
      failedToLoad: 'Failed to load subscriptions',
      failedToAssign: 'Failed to assign subscription',
      failedToExtend: 'Failed to extend subscription',
      failedToRevoke: 'Failed to revoke subscription',
      pleaseSelectUser: 'Bitte Benutzer auswaehlen',
      pleaseSelectGroup: 'Bitte waehlen Sie eine Gruppe',
      validityDaysRequired: 'Bitte geben Sie eine gueltige Anzahl von Tagen ein (mindestens 1)',
      revokeConfirm:
        "Are you sure you want to revoke the subscription for '{user}'? This action cannot be undone."
    },

    // Orders Management
    orders: {
      title: 'Bestellverwaltung',
      description: 'Manage all user orders',
      searchPlaceholder: 'Search order number...',
      allStatus: 'Alle Status',
      noOrdersYet: 'No orders yet',
      noOrdersDesc: 'Es gibt noch keine Bestelleintraege.',
      failedToLoad: 'Failed to load orders',
      userPrefix: 'User #{id}',
      rechargeSettings: 'Aufladeeinstellungen',
      tabs: {
        subscription: 'Abonnement-Bestellungen',
        recharge: 'Aufladebestellungen'

      },
      columns: {
        orderNo: 'Bestellnr.',
        user: 'Benutzer',
        group: 'Tarif',
        amount: 'Betrag',
        status: 'Status',
        payType: 'Zahlung',
        createdAt: 'Erstellt',
        paidAt: 'Paid At'
      },
      statusLabels: {
        pending: 'Ausstehend',
        paid: 'Bezahlt',
        expired: 'Abgelaufen',
        refunded: 'Refunded'
      }
    },

    // Referrals Management
    referrals: {
      title: 'Empfehlungsverwaltung',
      description: 'Alle Benutzer-Empfehlungseintraege anzeigen',
      searchPlaceholder: 'Search email...',
      noRecords: 'Keine Empfehlungseintraege',
      stats: {
        totalRecords: 'Total Referrals',
        totalReferrers: 'Empfehlungsgeber',
        pending: 'Ausstehend',
        rewarded: 'Rewarded',
        referrerPaid: 'Empfehlungsgeber-Belohnungen',
        inviteePaid: 'Invitee Rewards'
      },
      table: {
        referrer: 'Empfehlungsgeber',
        invitee: 'Invitee',
        status: 'Status',
        referrerReward: 'Empfehlungsgeber-Belohnung',
        inviteeReward: 'Invitee Reward',
        createdAt: 'Zeit'

      },
      status: {
        pending: 'Ausstehend',
        rewarded: 'Rewarded'
      }
    },

    // Recharge Orders Management
    rechargeOrders: {
      title: 'Aufladeverwaltung',
      description: 'Manage all user recharge orders',
      searchPlaceholder: 'Search order number, user...',
      allStatus: 'Alle Status',
      noOrders: 'Keine Aufladeeintraege',
      noOrdersDesc: 'Noch keine Aufladebestellungen.',
      failedToLoad: 'Failed to load recharge orders',
      user: 'Benutzer',
      tradeNo: 'Transaction No.',
      payMethod: 'Payment Method',
      paidAt: 'Paid At',
      totalOrders: 'Total Orders',
      totalAmount: 'Total Amount',
      totalCredit: 'Total Credit',
      averageMultiplier: 'Avg Multiplier'
    },

    // Accounts
    accounts: {
      title: 'Account Management',
      description: 'Manage AI platform accounts and credentials',
      createAccount: 'Konto erstellen',
      autoRefresh: 'Auto Refresh',
      enableAutoRefresh: 'Enable auto refresh',
      refreshInterval5s: '5 Sekunden',
      refreshInterval10s: '10 Sekunden',
      refreshInterval15s: '15 Sekunden',
      refreshInterval30s: '30 Sekunden',
      autoRefreshCountdown: 'Automatische Aktualisierung: {seconds}s',
      syncFromCrs: 'Von CRS synchronisieren',
      syncFromCrsTitle: 'Sync Accounts from CRS',
      syncFromCrsDesc:
        'Sync accounts from claude-relay-service (CRS) into this system (CRS is called server-to-server).',
      crsVersionRequirement: '⚠️ Note: CRS version must be ≥ v1.1.240 to support this feature',
      crsBaseUrl: 'CRS Basis-URL',
      crsBaseUrlPlaceholder: 'e.g. http://127.0.0.1:3000',
      crsUsername: 'Benutzername',
      crsPassword: 'Passwort',
      syncProxies: 'Auch Proxys synchronisieren (Abgleich nach Host/Port/Auth oder erstellen)',
      syncNow: 'Jetzt synchronisieren',
      syncing: 'Wird synchronisiert...',
      syncMissingFields: 'Bitte Basis-URL, Benutzername und Passwort ausfuellen',
      syncResult: 'Sync Result',
      syncResultSummary: 'Created {created}, updated {updated}, skipped {skipped}, failed {failed}',
      syncErrors: 'Errors / Skipped Details',
      syncCompleted: 'Synchronisierung abgeschlossen: {created} erstellt, {updated} aktualisiert',
      syncCompletedWithErrors:
        'Sync completed with errors: failed {failed} (created {created}, updated {updated})',
      syncFailed: 'Synchronisierung fehlgeschlagen',
      editAccount: 'Konto bearbeiten',
      deleteAccount: 'Konto loeschen',
      searchAccounts: 'Search accounts...',
      notes: 'Notizen',
      notesPlaceholder: 'Notizen eingeben',
      notesHint: 'Notizen sind optional',
      allPlatforms: 'All Platforms',
      allTypes: 'All Types',
      allStatus: 'Alle Status',
      oauthType: 'OAuth',
      setupToken: 'Setup Token',
      apiKey: 'API-Schluessel',

      // Schedulable toggle
      schedulable: 'Planbar',
      schedulableHint: 'Enable to include this account in API request scheduling',
      schedulableEnabled: 'Planung aktiviert',
      schedulableDisabled: 'Planung deaktiviert',
      failedToToggleSchedulable: 'Failed to toggle scheduling status',
      allGroups: '{count} Gruppen gesamt',
      platforms: {
        anthropic: 'Anthropic',
        claude: 'Claude',
        openai: 'OpenAI',
        gemini: 'Gemini',
        antigravity: 'Antigravity'

      },
      types: {
        oauth: 'OAuth',
        chatgptOauth: 'ChatGPT OAuth',
        responsesApi: 'Responses API',
        googleOauth: 'Google OAuth',
        codeAssist: 'Code Assist',
        antigravityOauth: 'Antigravity OAuth',
        api_key: 'API-Schluessel',
        cookie: 'Cookie',
        upstream: 'Upstream',
        upstreamDesc: 'Verbindung ueber Basis-URL + API-Schluessel'

      },
      status: {
        active: 'Aktiv',
        inactive: 'Inaktiv',
        error: 'Fehler',
        cooldown: 'Cooldown',
        paused: 'Paused',
        limited: 'Limited',
        rateLimited: 'Ratenbegrenzt',
        overloaded: 'Ueberlastet',
        tempUnschedulable: 'Temp Unschedulable',
        rateLimitedUntil: 'Rate limited until {time}',
        scopeRateLimitedUntil: '{scope} ratenbegrenzt bis {time}',
        overloadedUntil: 'Overloaded until {time}',
        viewTempUnschedDetails: 'Details der temporaeren Nichtplanbarkeit anzeigen'

      },
      columns: {
        name: 'Name',
        platformType: 'Platform/Type',
        platform: 'Plattform',
        type: 'Typ',
        capacity: 'Kapazitaet',
        notes: 'Notizen',
        priority: 'Prioritaet',
        billingRateMultiplier: 'Abrechnungsrate',
        weight: 'Gewicht',
        status: 'Status',
        schedulable: 'Planbar',
        todayStats: 'Today Stats',
        groups: 'Gruppen',
        usageWindows: 'Usage Windows',
        proxy: 'Proxy',
        lastUsed: 'Zuletzt verwendet',
        expiresAt: 'Laeuft ab am',
        actions: 'Aktionen'

      },
      // Capacity status tooltips
      capacity: {
        windowCost: {
          blocked: '5h-Fensterkosten ueberschritten, Kontoplanung pausiert',
          stickyOnly: '5h-Fensterkosten am Schwellenwert, nur persistente Sitzungen erlaubt',
          normal: '5h-Fensterkosten normal'

        },
        sessions: {
          full: 'Active sessions full, new sessions must wait (idle timeout: {idle} min)',
          normal: 'Active sessions normal (idle timeout: {idle} min)'
        }
      },
      tempUnschedulable: {
        title: 'Temp Unschedulable',
        statusTitle: 'Temp Unschedulable Status',
        hint: 'Disable accounts temporarily when error code and keyword both match.',
        notice: 'Regeln werden der Reihe nach ausgewertet und erfordern sowohl Fehlercode- als auch Schluesselwort-Uebereinstimmung.',
        addRule: 'Regel hinzufuegen',
        ruleOrder: 'Regelreihenfolge',
        ruleIndex: 'Regel #{index}',
        errorCode: 'Error Code',
        errorCodePlaceholder: 'z.B. 429',
        durationMinutes: 'Duration (minutes)',
        durationPlaceholder: 'z.B. 30',
        keywords: 'Schluesselwoerter',
        keywordsPlaceholder: 'e.g. overloaded, too many requests',
        keywordsHint: 'Schluesselwoerter durch Kommas trennen; jede Schluesselwort-Uebereinstimmung loest aus.',
        description: 'Beschreibung',
        descriptionPlaceholder: 'Optionale Notiz fuer diese Regel',
        rulesInvalid: 'Fuegen Sie mindestens eine Regel mit Fehlercode, Schluesselwoertern und Dauer hinzu.',
        viewDetails: 'Details der temporaeren Nichtplanbarkeit anzeigen',
        accountName: 'Konto',
        triggeredAt: 'Ausgeloest am',
        until: 'Bis',
        remaining: 'Verbleibend',
        matchedKeyword: 'Uebereinstimmendes Schluesselwort',
        errorMessage: 'Fehlerdetails',
        reset: 'Reset Status',
        resetSuccess: 'Temporaerer Nichtplanbar-Status zurueckgesetzt',
        resetFailed: 'Failed to reset temp unschedulable status',
        failedToLoad: 'Failed to load temp unschedulable status',
        notActive: 'Dieses Konto ist nicht temporaer nichtplanbar.',
        expired: 'Abgelaufen',
        remainingMinutes: 'Etwa {minutes} Minuten',
        remainingHours: 'Etwa {hours} Stunden',
        remainingHoursMinutes: 'Etwa {hours} Stunden {minutes} Minuten',
        presets: {
          overloadLabel: '529 Overloaded',
          overloadDesc: 'Overloaded - pause 60 minutes',
          rateLimitLabel: '429 Rate Limit',
          rateLimitDesc: 'Rate limited - pause 10 minutes',
          unavailableLabel: '503 Nicht verfuegbar',
          unavailableDesc: 'Nicht verfuegbar - 30 Minuten pausieren'

        }
      },
      clearRateLimit: 'Clear Rate Limit',
      testConnection: 'Verbindung testen',
      reAuthorize: 'Erneut autorisieren',
      refreshToken: 'Refresh Token',
      noAccountsYet: 'No accounts yet',
      createFirstAccount: 'Create your first account to start using AI services.',
      tokenRefreshed: 'Token erfolgreich aktualisiert',
      accountDeleted: 'Account deleted successfully',
      rateLimitCleared: 'Rate limit cleared successfully',
      bulkSchedulableEnabled: 'Successfully enabled scheduling for {count} account(s)',
      bulkSchedulableDisabled: 'Successfully disabled scheduling for {count} account(s)',
      bulkSchedulablePartial: 'Planung teilweise aktualisiert: {success} erfolgreich, {failed} fehlgeschlagen',
      bulkSchedulableResultUnknown: 'Massenplanungsergebnis unvollstaendig. Bitte erneut versuchen oder aktualisieren.',
      bulkActions: {
        selected: '{count} Konto/Konten ausgewaehlt',
        selectCurrentPage: 'Diese Seite auswaehlen',
        clear: 'Auswahl aufheben',
        edit: 'Massenbearbeitung',
        delete: 'Massenloeschung',
        enableScheduling: 'Enable Scheduling',
        disableScheduling: 'Disable Scheduling'
      },
      bulkEdit: {
        title: 'Bulk Edit Accounts',
        selectionInfo:
          '{count} account(s) selected. Only checked or filled fields will be updated; others stay unchanged.',
        baseUrlPlaceholder: 'https://api.anthropic.com or https://api.openai.com',
        baseUrlNotice: 'Gilt nur fuer API-Schluessel-Konten; leer lassen um bestehenden Wert beizubehalten',
        submit: 'Update Accounts',
        updating: 'Wird aktualisiert...',
        success: 'Updated {count} account(s)',
        partialSuccess: 'Teilweise aktualisiert: {success} erfolgreich, {failed} fehlgeschlagen',
        failed: 'Massenaktualisierung fehlgeschlagen',
        noSelection: 'Bitte Konten zum Bearbeiten auswaehlen',
        noFieldsSelected: 'Mindestens ein Feld zum Aktualisieren auswaehlen'

      },
      bulkDeleteTitle: 'Bulk Delete Accounts',
      bulkDeleteConfirm: 'Delete the selected {count} account(s)? This action cannot be undone.',
      bulkDeleteSuccess: 'Deleted {count} account(s)',
      bulkDeletePartial: 'Teilweise geloescht: {success} erfolgreich, {failed} fehlgeschlagen',
      bulkDeleteFailed: 'Massenloeschung fehlgeschlagen',
      resetStatus: 'Reset Status',
      statusReset: 'Account status reset successfully',
      failedToResetStatus: 'Failed to reset account status',
      failedToLoad: 'Failed to load accounts',
      failedToRefresh: 'Failed to refresh token',
      failedToDelete: 'Failed to delete account',
      dataImport: 'Daten importieren',
      dataExport: 'Daten exportieren',
      dataExportSelected: 'Auswahl exportieren',
      dataExportConfirmMessage: 'Kontodaten wirklich exportieren?',
      dataExportConfirm: 'Exportieren',
      dataExportIncludeProxies: 'Proxy-Konfiguration einschliessen',
      dataExported: 'Daten erfolgreich exportiert',
      dataExportFailed: 'Datenexport fehlgeschlagen',
      dataImportTitle: 'Kontodaten importieren',
      dataImportHint: 'Unterstuetzt den Import von Kontodaten im JSON-Format',
      dataImportWarning: 'Beim Import werden gleichnamige Konten ueberschrieben. Bitte mit Vorsicht vorgehen.',
      dataImportFile: 'Datei auswaehlen',
      dataImportSelectFile: 'Bitte waehlen Sie eine Datei zum Importieren',
      dataImportResult: 'Importergebnis',
      dataImportResultSummary: '{success} erfolgreich, {failed} fehlgeschlagen, {skipped} uebersprungen',
      dataImportErrors: 'Fehlerdetails',
      dataImporting: 'Importiere...',
      dataImportButton: 'Import starten',
      dataImportSuccess: 'Import abgeschlossen: {success} erfolgreich',
      dataImportCompletedWithErrors: 'Import abgeschlossen: {success} erfolgreich, {failed} fehlgeschlagen',
      dataImportParseFailed: 'Datei konnte nicht analysiert werden. Bitte ueberpruefen Sie das Format.',
      dataImportFailed: 'Datenimport fehlgeschlagen',
      failedToClearRateLimit: 'Failed to clear rate limit',
      deleteConfirm: "Moechten Sie '{name}' wirklich loeschen? Diese Aktion kann nicht rueckgaengig gemacht werden.",
      deleteConfirmMessage: "Moechten Sie das Konto '{name}' wirklich loeschen?",
      refreshCookie: 'Refresh Cookie',
      testAccount: 'Konto testen',
      form: {
        nameLabel: 'Kontoname',
        namePlaceholder: 'Enter account name',
        platformLabel: 'Plattform',
        selectPlatform: 'Select platform',
        typeLabel: 'Typ',
        selectType: 'Select type',
        credentialsLabel: 'Zugangsdaten',
        credentialsPlaceholder: 'Cookie oder API-Schluessel eingeben',
        priorityLabel: 'Prioritaet',
        priorityHint: 'Niedrigerer Wert = hoehere Prioritaet',
        weightLabel: 'Gewicht',
        weightHint: 'Gewichtswert fuer Lastverteilung',
        statusLabel: 'Status'

      },
      filters: {
        platform: 'Plattform',
        allPlatforms: 'All Platforms',
        type: 'Typ',
        allTypes: 'All Types',
        status: 'Status',
        allStatuses: 'Alle Status'

      },
      saving: 'Wird gespeichert...',
      refreshing: 'Refreshing...',
      noAccounts: 'No accounts',
      noAccountsDescription: 'Fuegen Sie KI-Plattform-Konten hinzu, um das API-Gateway zu verwenden.',
      accountCreatedSuccess: 'Account added successfully',
      accountUpdatedSuccess: 'Account updated successfully',
      accountDeletedSuccess: 'Account deleted successfully',
      cookieRefreshedSuccess: 'Cookie erfolgreich aktualisiert',
      testSuccess: 'Account test passed',
      failedToSave: 'Failed to save account',
      // Create/Edit Account Modal
      platform: 'Plattform',
      accountName: 'Kontoname',
      enterAccountName: 'Enter account name',
      accountType: 'Kontotyp',
      claudeCode: 'Claude Code',
      claudeConsole: 'Claude Console',
      oauthSetupToken: 'OAuth / Setup Token',
      addMethod: 'Hinzufuegemethode',
      setupTokenLongLived: 'Setup Token (Long-lived)',
      baseUrl: 'Basis-URL',
      baseUrlHint: 'Standard beibehalten fuer offizielle Anthropic API',
      apiKeyRequired: 'API-Schluessel *',
      apiKeyPlaceholder: 'sk-ant-api03-...',
      apiKeyHint: 'Ihr Claude Console API-Schluessel',

      // OpenAI specific hints
      openai: {
        baseUrlHint: 'Standard beibehalten fuer offizielle OpenAI API',
        apiKeyHint: 'Ihr OpenAI API-Schluessel'

      },
      modelRestriction: 'Model Restriction (Optional)',
      modelWhitelist: 'Model Whitelist',
      modelMapping: 'Model Mapping',
      selectAllowedModels: 'Erlaubte Modelle auswaehlen. Leer lassen, um alle Modelle zu unterstuetzen.',
      mapRequestModels:
        'Map request models to actual models. Left is the requested model, right is the actual model sent to API.',
      selectedModels: '{count} Modell(e) ausgewaehlt',
      supportsAllModels: '(unterstuetzt alle Modelle)',
      requestModel: 'Anfragemodell',
      actualModel: 'Actual model',
      addMapping: 'Zuordnung hinzufuegen',
      mappingExists: 'Mapping for {model} already exists',
      searchModels: 'Search models...',
      noMatchingModels: 'Keine passenden Modelle',
      fillRelatedModels: 'Verwandte Modelle ausfuellen',
      clearAllModels: 'Alle Modelle loeschen',
      customModelName: 'Benutzerdefinierter Modellname',
      enterCustomModelName: 'Benutzerdefinierten Modellnamen eingeben',
      addModel: 'Hinzufuegen',
      modelExists: 'Model already exists',
      modelCount: '{count} Modelle',
      customErrorCodes: 'Custom Error Codes',
      customErrorCodesHint: 'Planung nur fuer ausgewaehlte Fehlercodes stoppen',
      customErrorCodesWarning:
        'Only selected error codes will stop scheduling. Other errors will return 500.',
      customErrorCodes429Warning:
        '429 already has built-in rate limit handling. Adding it to custom error codes will disable the account instead of temporary rate limiting. Are you sure?',
      customErrorCodes529Warning:
        '529 already has built-in overload handling. Adding it to custom error codes will disable the account instead of temporary overload marking. Are you sure?',
      selectedErrorCodes: 'Ausgewaehlt',
      noneSelectedUsesDefault: 'None selected (uses default policy)',
      enterErrorCode: 'Fehlercode eingeben (100-599)',
      invalidErrorCode: 'Bitte geben Sie einen gueltigen HTTP-Fehlercode ein (100-599)',
      errorCodeExists: 'This error code is already selected',
      interceptWarmupRequests: 'Intercept Warmup Requests',
      interceptWarmupRequestsDesc:
        'When enabled, warmup requests like title generation will return mock responses without consuming upstream tokens',
      autoPauseOnExpired: 'Auto Pause On Expired',
      autoPauseOnExpiredDesc: 'Wenn aktiviert, wird die Planung des Kontos nach Ablauf automatisch pausiert',

      // Quota control (Anthropic OAuth/SetupToken only)
      quotaControl: {
        title: 'Kontingentkontrolle',
        hint: 'Only applies to Anthropic OAuth/Setup Token accounts',
        windowCost: {
          label: '5h Window Cost Limit',
          hint: 'Kontokostennutzung innerhalb des 5-Stunden-Fensters begrenzen',
          limit: 'Cost Threshold',
          limitPlaceholder: '50',
          limitHint: 'Account will not participate in new scheduling after reaching threshold',
          stickyReserve: 'Persistente Reserve',
          stickyReservePlaceholder: '10',
          stickyReserveHint: 'Zusaetzliche Reserve fuer persistente Sitzungen'

        },
        sessionLimit: {
          label: 'Sitzungsanzahl-Limit',
          hint: 'Limit the number of active concurrent sessions',
          maxSessions: 'Max. Sitzungen',
          maxSessionsPlaceholder: '3',
          maxSessionsHint: 'Maximum number of active concurrent sessions',
          idleTimeout: 'Idle Timeout',
          idleTimeoutPlaceholder: '5',
          idleTimeoutHint: 'Sessions will be released after idle timeout'
        },
        tlsFingerprint: {
          label: 'TLS-Fingerabdruck-Simulation',
          hint: 'Node.js/Claude Code Client TLS-Fingerabdruck simulieren'

        },
        sessionIdMasking: {
          label: 'Sitzungs-ID-Maskierung',
          hint: 'When enabled, fixes the session ID in metadata.user_id for 15 minutes, making upstream think requests come from the same session'
        }
      },
      expired: 'Abgelaufen',
      proxy: 'Proxy',
      noProxy: 'Kein Proxy',
      concurrency: 'Parallelitaet',
      priority: 'Prioritaet',
      priorityHint: 'Konten mit niedrigerem Wert werden zuerst verwendet',
      billingRateMultiplier: 'Billing Rate Multiplier',
      billingRateMultiplierHint: '>=0, 0 bedeutet kostenlos. Betrifft nur die Kontoabrechnung',
      expiresAt: 'Laeuft ab am',
      expiresAtHint: 'Leer lassen fuer kein Ablaufdatum',
      higherPriorityFirst: 'Niedrigerer Wert bedeutet hoehere Prioritaet',
      mixedScheduling: 'Use in /v1/messages',
      mixedSchedulingHint: 'Enable to participate in Anthropic/Gemini group scheduling',
      mixedSchedulingTooltip:
        '!! WARNING !! Antigravity Claude and Anthropic Claude cannot be used in the same context. If you have both Anthropic and Antigravity accounts, enabling this option will cause frequent 400 errors. When enabled, please use the group feature to isolate Antigravity accounts from Anthropic accounts. Make sure you understand this before enabling!!',
      creating: 'Wird erstellt...',
      updating: 'Wird aktualisiert...',
      accountCreated: 'Account created successfully',
      accountUpdated: 'Account updated successfully',
      failedToCreate: 'Failed to create account',
      failedToUpdate: 'Failed to update account',
      mixedChannelWarningTitle: 'Mixed Channel Warning',
      mixedChannelWarning: 'Warning: Group "{groupName}" contains both {currentPlatform} and {otherPlatform} accounts. Mixing different channels may cause thinking block signature validation issues, which will fallback to non-thinking mode. Are you sure you want to continue?',
      pleaseEnterAccountName: 'Bitte Kontoname eingeben',
      pleaseEnterApiKey: 'Bitte API-Schluessel eingeben',
      apiKeyIsRequired: 'API-Schluessel ist erforderlich',
      leaveEmptyToKeep: 'Leave empty to keep current key',
      // Upstream type
      upstream: {
        baseUrl: 'Upstream Basis-URL',
        baseUrlHint: 'Die Adresse des Upstream-Antigravity-Dienstes, z.B. https://s.konstants.xyz',
        apiKey: 'Upstream API-Schluessel',
        apiKeyHint: 'API-Schluessel fuer den Upstream-Dienst',
        pleaseEnterBaseUrl: 'Bitte Upstream Basis-URL eingeben',
        pleaseEnterApiKey: 'Bitte Upstream API-Schluessel eingeben'

      },
      // OAuth flow
      oauth: {
        title: 'Claude Account Authorization',
        authMethod: 'Autorisierungsmethode',
        manualAuth: 'Manuelle Autorisierung',
        cookieAutoAuth: 'Cookie-Auto-Auth',
        cookieAutoAuthDesc:
          'Use claude.ai sessionKey to automatically complete OAuth authorization without manually opening browser.',
        sessionKey: 'sessionKey',
        keysCount: '{count} Schluessel',
        batchCreateAccounts: '{count} Konten werden in einem Stapel erstellt',
        sessionKeyPlaceholder:
          'One sessionKey per line, e.g.:\nsk-ant-sid01-xxxxx...\nsk-ant-sid01-yyyyy...',
        sessionKeyPlaceholderSingle: 'sk-ant-sid01-xxxxx...',
        howToGetSessionKey: 'Wie man den sessionKey erhaelt',
        step1: 'Login to claude.ai in your browser',
        step2: 'Druecken Sie F12, um die Entwicklertools zu oeffnen',
        step3: 'Gehen Sie zum Tab Application',
        step4: 'Find Cookies → https://claude.ai',
        step5: 'Finden Sie die Zeile mit dem Schluessel sessionKey',
        step6: 'Copy the Value',
        sessionKeyFormat: 'sessionKey beginnt normalerweise mit sk-ant-sid01-',
        startAutoAuth: 'Auto-Auth starten',
        authorizing: 'Autorisierung laeuft...',
        followSteps: 'Befolgen Sie diese Schritte, um Ihr Claude-Konto zu autorisieren:',
        step1GenerateUrl: 'Klicken Sie auf die Schaltflaeche unten, um die Autorisierungs-URL zu generieren',
        generateAuthUrl: 'Generate Auth URL',
        generating: 'Wird generiert...',
        regenerate: 'Neu generieren',
        step2OpenUrl: 'Oeffnen Sie die URL in Ihrem Browser und schliessen Sie die Autorisierung ab',
        openUrlDesc:
          'Open the authorization URL in a new tab, log in to your Claude account and authorize.',
        proxyWarning:
          'Note: If you configured a proxy, make sure your browser uses the same proxy to access the authorization page.',
        step3EnterCode: 'Autorisierungscode eingeben',
        authCodeDesc:
          'After authorization is complete, the page will display an Authorization Code. Copy and paste it below:',
        authCode: 'Autorisierungscode',
        authCodePlaceholder: 'Autorisierungscode von der Claude-Seite einfuegen...',
        authCodeHint: 'Den von der Claude-Seite kopierten Autorisierungscode einfuegen',
        completeAuth: 'Autorisierung abschliessen',
        verifying: 'Wird verifiziert...',
        pleaseEnterSessionKey: 'Bitte geben Sie mindestens einen gueltigen sessionKey ein',
        authFailed: 'Autorisierung fehlgeschlagen',
        cookieAuthFailed: 'Cookie-Autorisierung fehlgeschlagen',
        keyAuthFailed: 'Schluessel {index}: {error}',
        successCreated: 'Successfully created {count} account(s)',
        // OpenAI specific
        openai: {
          title: 'OpenAI Account Authorization',
          followSteps: 'Befolgen Sie diese Schritte, um die OpenAI-Kontoautorisierung abzuschliessen:',
          step1GenerateUrl: 'Klicken Sie auf die Schaltflaeche unten, um die Autorisierungs-URL zu generieren',
          generateAuthUrl: 'Generate Auth URL',
          step2OpenUrl: 'Oeffnen Sie die URL in Ihrem Browser und schliessen Sie die Autorisierung ab',
          openUrlDesc:
            'Open the authorization URL in a new tab, log in to your OpenAI account and authorize.',
          importantNotice:
            'Important: The page may take a while to load after authorization. Please wait patiently. When the browser address bar changes to http://localhost..., the authorization is complete.',
          step3EnterCode: 'Autorisierungs-URL oder Code eingeben',
          authCodeDesc:
            'After authorization is complete, when the page URL becomes http://localhost:xxx/auth/callback?code=...:',
          authCode: 'Autorisierungs-URL oder Code',
          authCodePlaceholder:
            'Option 1: Copy the complete URL\n(http://localhost:xxx/auth/callback?code=...)\nOption 2: Copy only the code parameter value',
          authCodeHint:
            'You can copy the entire URL or just the code parameter value, the system will auto-detect'
        },
        // Gemini specific
	        gemini: {
	          title: 'Gemini Account Authorization',
	          followSteps: 'Befolgen Sie diese Schritte, um Ihr Gemini-Konto zu autorisieren:',
	          step1GenerateUrl: 'Generate the authorization URL',
	          generateAuthUrl: 'Generate Auth URL',
	          projectIdLabel: 'Projekt-ID (optional)',
	          projectIdPlaceholder: 'e.g. my-gcp-project or cloud-ai-companion-xxxxx',
	          projectIdHint:
	            'Leave empty to auto-detect after code exchange. If auto-detection fails, fill it in and re-generate the auth URL to try again.',
	          howToGetProjectId: 'Wie man erhaelt',
	          step2OpenUrl: 'Oeffnen Sie die URL in Ihrem Browser und schliessen Sie die Autorisierung ab',
	          openUrlDesc:
	            'Open the authorization URL in a new tab, log in to your Google account and authorize.',
	          step3EnterCode: 'Autorisierungs-URL oder Code eingeben',
	          authCodeDesc:
	            'After authorization, copy the callback URL (recommended) or just the code and paste it below.',
	          authCode: 'Callback-URL oder Code',
	          authCodePlaceholder:
	            'Option 1 (recommended): Paste the callback URL\nOption 2: Paste only the code value',
	          authCodeHint: 'Das System extrahiert automatisch Code/Status aus der URL.',
          redirectUri: 'Weiterleitungs-URI',
          redirectUriHint:
            'This must be configured in your Google OAuth client and must match exactly.',
          confirmRedirectUri:
            'I have configured this Redirect URI in the Google OAuth client (must match exactly)',
	          invalidRedirectUri: 'Weiterleitungs-URI muss eine gueltige http(s)-URL sein',
	          redirectUriNotConfirmed: 'Bitte bestaetigen Sie, dass die Weiterleitungs-URI korrekt konfiguriert ist',
	          missingRedirectUri: 'Fehlende Weiterleitungs-URI',
	          failedToGenerateUrl: 'Failed to generate Gemini auth URL',
	          missingExchangeParams: 'Fehlender Auth-Code, Sitzungs-ID oder Status',
	          failedToExchangeCode: 'Failed to exchange Gemini auth code',
	          missingProjectId: 'GCP Project ID retrieval failed: Your Google account is not linked to an active GCP project. Please activate GCP and bind a credit card in Google Cloud Console, or manually enter the Project ID during authorization.',
	          modelPassthrough: 'Gemini Model Passthrough',
	          modelPassthroughDesc:
	            'All model requests are forwarded directly to the Gemini API without model restrictions or mappings.',
	          stateWarningTitle: 'Hinweis',
	          stateWarningDesc: 'Recommended: paste the full callback URL (includes code & state).',
	          oauthTypeLabel: 'OAuth Type',
          needsProjectId: 'Eingebautes OAuth (Code Assist)',
          needsProjectIdDesc: 'Erfordert GCP-Projekt und Projekt-ID',
          noProjectIdNeeded: 'Benutzerdefiniertes OAuth (AI Studio)',
          noProjectIdNeededDesc: 'Erfordert vom Admin konfigurierten OAuth-Client',
	          aiStudioNotConfiguredShort: 'Nicht konfiguriert',
	          aiStudioNotConfiguredTip:
	            'AI Studio OAuth is not configured: set GEMINI_OAUTH_CLIENT_ID / GEMINI_OAUTH_CLIENT_SECRET and add Redirect URI: http://localhost:1455/auth/callback (Consent screen scopes must include https://www.googleapis.com/auth/generative-language.retriever)',
	          aiStudioNotConfigured:
	            'AI Studio OAuth is not configured: set GEMINI_OAUTH_CLIENT_ID / GEMINI_OAUTH_CLIENT_SECRET and add Redirect URI: http://localhost:1455/auth/callback'
	        },
        // Antigravity specific
        antigravity: {
          title: 'Antigravity Account Authorization',
          followSteps: 'Befolgen Sie diese Schritte, um Ihr Antigravity-Konto zu autorisieren:',
          step1GenerateUrl: 'Generate the authorization URL',
          generateAuthUrl: 'Generate Auth URL',
          step2OpenUrl: 'Oeffnen Sie die URL in Ihrem Browser und schliessen Sie die Autorisierung ab',
          openUrlDesc: 'Oeffnen Sie die Autorisierungs-URL in einem neuen Tab, melden Sie sich bei Ihrem Google-Konto an und autorisieren Sie.',
          importantNotice:
            'Important: The page may take a while to load after authorization. Please wait patiently. When the browser address bar shows http://localhost..., authorization is complete.',
          step3EnterCode: 'Autorisierungs-URL oder Code eingeben',
          authCodeDesc:
            'After authorization, when the page URL becomes http://localhost:xxx/auth/callback?code=...:',
          authCode: 'Autorisierungs-URL oder Code',
          authCodePlaceholder:
            'Option 1: Copy the complete URL\n(http://localhost:xxx/auth/callback?code=...)\nOption 2: Copy only the code parameter value',
          authCodeHint: 'Sie koennen die gesamte URL oder nur den Code-Parameterwert kopieren, das System erkennt es automatisch',
          failedToGenerateUrl: 'Failed to generate Antigravity auth URL',
          missingExchangeParams: 'Fehlender Code, Sitzungs-ID oder Status',
          failedToExchangeCode: 'Failed to exchange Antigravity auth code'
        }
	      },
      // Gemini specific (platform-wide)
      gemini: {
        helpButton: 'Hilfe',
        helpDialog: {
          title: 'Gemini Usage Guide',
          apiKeySection: 'API-Schluessel-Links'

        },
        modelPassthrough: 'Gemini Model Passthrough',
        modelPassthroughDesc:
          'All model requests are forwarded directly to the Gemini API without model restrictions or mappings.',
        baseUrlHint: 'Standard beibehalten fuer offizielle Gemini API',
        apiKeyHint: 'Ihr Gemini API-Schluessel (beginnt mit AIza)',
        tier: {
          label: 'Account Tier',
          hint: 'Tip: The system will try to auto-detect the tier first; if auto-detection is unavailable or fails, your selected tier is used as a fallback (simulated quota).',
          aiStudioHint:
            'AI Studio quotas are per-model (Pro/Flash are limited independently). If billing is enabled, choose Pay-as-you-go.',
          googleOne: {
            free: 'Google One Kostenlos',
            pro: 'Google One Pro',
            ultra: 'Google One Ultra'

          },
          gcp: {
            standard: 'GCP Standard',
            enterprise: 'GCP Enterprise'

          },
          aiStudio: {
            free: 'Google AI Kostenlos',
            paid: 'Google AI Pay-as-you-go'

          }
        },
        accountType: {
          oauthTitle: 'OAuth (Gemini)',
          oauthDesc: 'Autorisieren Sie sich mit Ihrem Google-Konto und waehlen Sie einen OAuth-Typ.',
          apiKeyTitle: 'API-Schluessel (AI Studio)',
          apiKeyDesc: 'Schnellste Einrichtung. Verwenden Sie einen AIza API-Schluessel.',
          apiKeyNote:
            'Best for light testing. Free tier has strict rate limits and data may be used for training.',
          apiKeyLink: 'API-Schluessel holen',
          quotaLink: 'Quota guide'
        },
        oauthType: {
          builtInTitle: 'Eingebautes OAuth (Gemini CLI / Code Assist)',
          builtInDesc: 'Verwendet die eingebaute Google-Client-ID. Keine Admin-Konfiguration erforderlich.',
          builtInRequirement: 'Erfordert ein GCP-Projekt und eine Projekt-ID.',
          gcpProjectLink: 'Create project',
          customTitle: 'Benutzerdefiniertes OAuth (AI Studio OAuth)',
          customDesc: 'Verwendet vom Admin konfigurierten OAuth-Client fuer Organisationsverwaltung.',
          customRequirement: 'Admin muss Client-ID konfigurieren und Sie als Testbenutzer hinzufuegen.',
          badges: {
            recommended: 'Empfohlen',
            highConcurrency: 'Hohe Parallelitaet',
            noAdmin: 'Kein Admin-Setup',
            orgManaged: 'Org-verwaltet',
            adminRequired: 'Admin erforderlich'

          }
        },
        setupGuide: {
          title: 'Gemini Setup Checklist',
          checklistTitle: 'Checkliste',
          checklistItems: {
            usIp: 'Verwenden Sie eine US-IP und stellen Sie sicher, dass Ihr Kontoland auf US eingestellt ist.',
            age: 'Account must be 18+.'
          },
          activationTitle: 'Ein-Klick-Aktivierung',
          activationItems: {
            geminiWeb: 'Activate Gemini Web to avoid User not initialized.',
            gcpProject: 'Aktivieren Sie ein GCP-Projekt und erhalten Sie die Projekt-ID fuer Code Assist.'

          },
          links: {
            countryCheck: 'Landzuordnung pruefen',
            geminiWebActivation: 'Gemini Web aktivieren',
            gcpProject: 'GCP-Konsole oeffnen'

          }
        },
        quotaPolicy: {
          title: 'Gemini Quota & Limit Policy (Reference)',
          note: 'Note: Gemini does not provide an official quota inquiry API. The "Daily Quota" shown here is an estimate simulated by the system based on account tiers for scheduling reference only. Please refer to official Google errors for actual limits.',
          columns: {
            channel: 'Auth-Kanal',
            account: 'Account Status',
            limits: 'Limit-Richtlinie',
            docs: 'Official Docs'
          },
          docs: {
            codeAssist: 'Code Assist Quotas',
            aiStudio: 'AI Studio Pricing',
            vertex: 'Vertex AI Quotas'
          },
          simulatedNote: 'Simuliertes Kontingent, nur zur Referenz',
          rows: {
            googleOne: {
              channel: 'Google One OAuth (Individuals / Code Assist for Individuals)',
              limitsFree: 'Shared pool: 1000 RPD / 60 RPM',
              limitsPro: 'Shared pool: 1500 RPD / 120 RPM',
              limitsUltra: 'Shared pool: 2000 RPD / 120 RPM'
            },
            gcp: {
              channel: 'GCP Code Assist OAuth (Enterprise)',
              limitsStandard: 'Shared pool: 1500 RPD / 120 RPM',
              limitsEnterprise: 'Shared pool: 2000 RPD / 120 RPM'
            },
            cli: {
              channel: 'Gemini CLI (Official Google Login / Code Assist)',
              free: 'Free Google Account',
              premium: 'Google One AI Premium',
              limitsFree: 'RPD ~1000; RPM ~60 (soft)',
              limitsPremium: 'RPD ~1500+; RPM ~60+ (priority queue)'
            },
            gcloud: {
              channel: 'GCP Code Assist (gcloud auth)',
              account: 'No Code Assist subscription',
              limits: 'RPD ~1000; RPM ~60 (preview)'
            },
            aiStudio: {
              channel: 'AI Studio API Key / OAuth',
              free: 'No billing (free tier)',
              paid: 'Billing enabled (pay-as-you-go)',
              limitsFree: 'RPD 50; RPM 2 (Pro) / 15 (Flash)',
              limitsPaid: 'RPD unlimited; RPM 1000 (Pro) / 2000 (Flash) (per model)'
            },
            customOAuth: {
              channel: 'Custom OAuth Client (GCP)',
              free: 'Project not billed',
              paid: 'Project billed',
              limitsFree: 'RPD 50; RPM 2 (project quota)',
              limitsPaid: 'RPD unlimited; RPM 1000+ (project quota)'
            }
          }
        },
        rateLimit: {
          ok: 'Nicht ratenbegrenzt',
          unlimited: 'Unbegrenzt',
          limited: 'Rate limited {time}',
          now: 'jetzt'

        }
      },
      // Re-Auth Modal
      reAuthorizeAccount: 'Re-Authorize Account',
      claudeCodeAccount: 'Claude Code Account',
      openaiAccount: 'OpenAI Account',
      geminiAccount: 'Gemini Account',
      antigravityAccount: 'Antigravity Account',
      inputMethod: 'Input Method',
      reAuthorizedSuccess: 'Account re-authorized successfully',
      // Test Modal
      testAccountConnection: 'Test Account Connection',
      account: 'Konto',
      readyToTest: 'Ready to test. Click "Start Test" to begin...',
      connectingToApi: 'Verbindung zur API wird hergestellt...',
      testCompleted: 'Test erfolgreich abgeschlossen!',
      testFailed: 'Test fehlgeschlagen',
      connectedToApi: 'Mit API verbunden',
      usingModel: 'Verwendetes Modell: {model}',
      sendingTestMessage: 'Sende Testnachricht: "hi"',
      response: 'Response:',
      startTest: 'Start Test',
      testing: 'Wird getestet...',
      retry: 'Wiederholen',
      copyOutput: 'Copy output',
      outputCopied: 'Output copied',
      startingTestForAccount: 'Test fuer Konto wird gestartet: {name}',
      testAccountTypeLabel: 'Account type: {type}',
      selectTestModel: 'Select Test Model',
      testModel: 'Testmodell',
      testPrompt: 'Prompt: "hi"',
      // Stats Modal
      viewStats: 'View Stats',
      usageStatistics: 'Usage Statistics',
      last30DaysUsage: 'Nutzungsstatistik der letzten 30 Tage (basierend auf tatsaechlichen Nutzungstagen)',
      stats: {
        totalCost: '30-Day Total Cost',
        accumulatedCost: 'Kumulierte Kosten',
        standardCost: 'Standard',
        totalRequests: '30-Day Total Requests',
        totalCalls: 'Total API calls',
        avgDailyCost: 'Daily Avg Cost',
        basedOnActualDays: 'Basierend auf {days} tatsaechlichen Nutzungstagen',
        avgDailyRequests: 'Daily Avg Requests',
        avgDailyUsage: 'Durchschnittliche taegliche Nutzung',
        todayOverview: 'Today Overview',
        cost: 'Kosten',
        requests: 'Anfragen',
        tokens: 'Tokens',
        highestCostDay: 'Highest Cost Day',
        highestRequestDay: 'Tag mit den meisten Anfragen',
        date: 'Datum',
        accumulatedTokens: 'Kumulierte Tokens',
        totalTokens: '30-Day Total',
        dailyAvgTokens: 'Daily Average',
        performance: 'Performance',
        avgResponseTime: 'Avg Response',
        daysActive: 'Days Active',
        recentActivity: 'Recent Activity',
        todayRequests: 'Today Requests',
        todayTokens: 'Today Tokens',
        todayCost: 'Today Cost',
        usageTrend: '30-Day Cost & Request Trend',
        noData: 'Keine Nutzungsdaten fuer dieses Konto verfuegbar'

      },
      usageWindow: {
        statsTitle: '5-Hour Window Usage Statistics',
        statsTitleDaily: 'Daily Usage Statistics',
        geminiProDaily: 'Pro',
        geminiFlashDaily: 'Flash',
        gemini3Pro: 'G3P',
        gemini3Flash: 'G3F',
        gemini3Image: 'G3I',
        claude45: 'C4.5'
      },
      tier: {
        free: 'Kostenlos',
        pro: 'Pro',
        ultra: 'Ultra',
        aiPremium: 'AI Premium',
        standard: 'Standard',
        basic: 'Basis',
        personal: 'Persoenlich',
        unlimited: 'Unbegrenzt'

      },
      ineligibleWarning:
        'This account is not eligible for Antigravity, but API forwarding still works. Use at your own risk.'
    },

    // Proxies
    proxies: {
      title: 'Proxy Management',
      description: 'Manage proxy servers for accounts',
      createProxy: 'Proxy erstellen',
      editProxy: 'Proxy bearbeiten',
      deleteProxy: 'Proxy loeschen',
      searchProxies: 'Search proxies...',
      allProtocols: 'Alle Protokolle',
      allStatus: 'Alle Status',
      protocols: {
        http: 'HTTP',
        https: 'HTTPS',
        socks5: 'SOCKS5',
        socks5h: 'SOCKS5H (Remote DNS)'

      },
      columns: {
        name: 'Name',
        protocol: 'Protokoll',
        address: 'Adresse',
        location: 'Standort',
        status: 'Status',
        accounts: 'Konten',
        latency: 'Latenz',
        actions: 'Aktionen',
        nameLabel: 'Name',
        namePlaceholder: 'Enter proxy name',
        protocolLabel: 'Protokoll',
        selectProtocol: 'Select protocol',
        hostLabel: 'Host',
        hostPlaceholder: 'Hostadresse eingeben',
        portLabel: 'Port',
        portPlaceholder: 'Port eingeben',
        usernameLabel: 'Username (Optional)',
        usernamePlaceholder: 'Benutzername eingeben',
        passwordLabel: 'Password (Optional)',
        passwordPlaceholder: 'Passwort eingeben',
        priorityLabel: 'Prioritaet',
        statusLabel: 'Status'

      },
      testConnection: 'Verbindung testen',
      batchTest: 'Test All Proxies',
      testFailed: 'Fehlgeschlagen',
      latencyFailed: 'Verbindung fehlgeschlagen',
      batchTestEmpty: 'No proxies available for testing',
      batchTestDone: 'Massentest fuer {count} Proxys abgeschlossen',
      batchTestFailed: 'Massentest fehlgeschlagen',
      batchDeleteAction: 'Loeschen',
      batchDelete: 'Massenloeschung',
      batchDeleteConfirm: 'Delete {count} selected proxies? In-use ones will be skipped.',
      batchDeleteDone: 'Deleted {deleted} proxies, skipped {skipped}',
      batchDeleteSkipped: '{skipped} Proxys uebersprungen',
      batchDeleteFailed: 'Massenloeschung fehlgeschlagen',
      deleteBlockedInUse: 'Dieser Proxy wird verwendet und kann nicht geloescht werden',
      accountsTitle: 'Accounts using this IP',
      accountsEmpty: 'No accounts are using this proxy',
      accountsFailed: 'Failed to load accounts list',
      accountName: 'Konto',
      accountPlatform: 'Plattform',
      accountNotes: 'Notizen',
      name: 'Name',
      protocol: 'Protokoll',
      host: 'Host',
      port: 'Port',
      username: 'Username (Optional)',
      password: 'Password (Optional)',
      status: 'Status',
      enterProxyName: 'Enter proxy name',
      leaveEmptyToKeep: 'Leave empty to keep current',
      optionalAuth: 'Optionale Authentifizierung',
      form: {
        hostPlaceholder: 'proxy.beispiel.de',
        portPlaceholder: '8080'

      },
      noProxiesYet: 'No proxies yet',
      createFirstProxy: 'Create your first proxy to route traffic through it.',
      // Batch import
      standardAdd: 'Standard hinzufuegen',
      batchAdd: 'Schnell hinzufuegen',
      batchInput: 'Proxy-Liste',
      batchInputPlaceholder:
        "Enter one proxy per line in the following formats:\nsocks5://user:pass{'@'}192.168.1.1:1080\nhttp://192.168.1.1:8080\nhttps://user:pass{'@'}proxy.example.com:443",
      batchInputHint:
        "Supports http, https, socks5 protocols. Format: protocol://[user:pass{'@'}]host:port",
      parsedCount: '{count} gueltig',
      invalidCount: '{count} ungueltig',
      duplicateCount: '{count} Duplikat(e)',
      importing: 'Importing...',
      importProxies: 'Import {count} proxies',
      batchImportSuccess: 'Successfully imported {created} proxies, skipped {skipped} duplicates',
      batchImportAllSkipped: 'All {skipped} proxies already exist, skipped import',
      failedToImport: 'Failed to batch import',
      dataImport: 'Daten importieren',
      dataExport: 'Daten exportieren',
      dataExportSelected: 'Auswahl exportieren',
      dataExportConfirmMessage: 'Proxy-Daten wirklich exportieren?',
      dataExportConfirm: 'Exportieren',
      dataExported: 'Daten erfolgreich exportiert',
      dataExportFailed: 'Datenexport fehlgeschlagen',
      dataImportTitle: 'Proxy-Daten importieren',
      dataImportHint: 'Unterstuetzt den Import von Proxy-Daten im JSON-Format',
      dataImportWarning: 'Beim Import werden gleichnamige Proxys ueberschrieben. Bitte mit Vorsicht vorgehen.',
      dataImportFile: 'Datei auswaehlen',
      dataImportSelectFile: 'Bitte waehlen Sie eine Datei zum Importieren',
      dataImportResult: 'Importergebnis',
      dataImportResultSummary: '{success} erfolgreich, {failed} fehlgeschlagen, {skipped} uebersprungen',
      dataImportErrors: 'Fehlerdetails',
      dataImporting: 'Importiere...',
      dataImportButton: 'Import starten',
      dataImportSuccess: 'Import abgeschlossen: {success} erfolgreich',
      dataImportCompletedWithErrors: 'Import abgeschlossen: {success} erfolgreich, {failed} fehlgeschlagen',
      dataImportParseFailed: 'Datei konnte nicht analysiert werden. Bitte ueberpruefen Sie das Format.',
      dataImportFailed: 'Datenimport fehlgeschlagen',
      // Other messages
      creating: 'Wird erstellt...',
      updating: 'Wird aktualisiert...',
      proxyCreated: 'Proxy erfolgreich erstellt',
      proxyUpdated: 'Proxy erfolgreich aktualisiert',
      proxyDeleted: 'Proxy erfolgreich geloescht',
      proxyWorking: 'Proxy funktioniert!',
      proxyWorkingWithLatency: 'Proxy funktioniert! Latenz: {latency}ms',
      proxyTestFailed: 'Proxy-Test fehlgeschlagen',
      failedToLoad: 'Failed to load proxies',
      failedToCreate: 'Failed to create proxy',
      failedToUpdate: 'Failed to update proxy',
      failedToDelete: 'Failed to delete proxy',
      failedToTest: 'Failed to test proxy',
      nameRequired: 'Bitte Proxy-Name eingeben',
      hostRequired: 'Bitte Hostadresse eingeben',
      portInvalid: 'Port muss zwischen 1-65535 liegen',
      deleteConfirm:
        "Are you sure you want to delete '{name}'? Accounts using this proxy will have their proxy removed.",
      deleteConfirmMessage: "Moechten Sie den Proxy '{name}' wirklich loeschen?",
      testProxy: 'Proxy testen',
      filters: {
        protocol: 'Protokoll',
        allProtocols: 'Alle Protokolle',
        status: 'Status',
        allStatuses: 'Alle Status'

      },
      saving: 'Wird gespeichert...',
      testing: 'Wird getestet...',
      noProxies: 'No proxies',
      noProxiesDescription: 'Proxy-Server hinzufuegen, um die API-Zugangsstabilitaet zu verbessern.',
      proxyCreatedSuccess: 'Proxy erfolgreich hinzugefuegt',
      proxyUpdatedSuccess: 'Proxy erfolgreich aktualisiert',
      proxyDeletedSuccess: 'Proxy erfolgreich geloescht',
      testSuccess: 'Proxy-Test bestanden',
      failedToSave: 'Failed to save proxy'
    },

    // Channel Management
    channels: {
      title: 'Channel Management',
      description: 'Kanaele konfigurieren und Guthaben abfragen',
      createChannel: 'Create Channel',
      editChannel: 'Edit Channel',
      deleteChannel: 'Delete Channel',
      searchChannels: 'Search channels...',
      allPlatforms: 'All Platforms',
      allStatus: 'Alle Status',
      noPlatform: 'Keine Plattform',
      other: 'Andere',
      noChannels: 'No channels yet',
      name: 'Kanalname',
      namePlaceholder: 'Enter channel name',
      descriptionLabel: 'Beschreibung',
      descriptionPlaceholder: 'Kanalbeschreibung eingeben',
      platform: 'Plattform',
      selectPlatform: 'Select platform',
      status: {
        active: 'Aktiv',
        inactive: 'Inaktiv',
        error: 'Fehler'

      },
      balanceConfig: 'Balance Query Configuration',
      balanceUrl: 'Abfrage-URL',
      balanceUrlPlaceholder: 'https://api.example.com/balance',
      balanceMethod: 'Anfragemethode',
      balanceUnit: 'Balance Unit',
      balancePath: 'Balance Field Path',
      balancePathPlaceholder: 'data.balance',
      balanceHeaders: 'Anfrage-Header',
      balanceHeadersPlaceholder: '"Authorization": "Bearer sk-xxx"',
      balanceHeadersHint: 'Anfrage-Header im JSON-Format eingeben',
      balanceBody: 'Anfrage-Body',
      balanceBodyPlaceholder: '"key": "value"',
      lastUpdated: 'Aktualisiert',
      refreshBalance: 'Refresh Balance',
      balanceRefreshed: 'Balance refreshed',
      balanceCheckFailed: 'Balance check failed',
      balanceCheckError: 'Balance check error',
      invalidHeadersJson: 'Ungueltiges JSON-Format fuer Header',
      loadError: 'Failed to load channels',
      createSuccess: 'Kanal erfolgreich erstellt',
      createError: 'Failed to create channel',
      updateSuccess: 'Kanal erfolgreich aktualisiert',
      updateError: 'Failed to update channel',
      deleteSuccess: 'Kanal erfolgreich geloescht',
      deleteError: 'Failed to delete channel',
      deleteConfirmMessage: 'Moechten Sie den Kanal "{name}" wirklich loeschen?',
      importCurl: 'Import cURL',
      curlCommand: 'cURL Command',
      curlPlaceholder: 'Paste cURL command from browser dev tools...',
      curlImportHint: 'Copy cURL command from browser dev tools (Network panel) to auto-parse URL and headers.',
      curlEmpty: 'Bitte geben Sie einen cURL-Befehl ein',
      parseCurl: 'Analysieren & Ausfuellen',
      autoRefresh: 'Auto Refresh',
      balance: 'Guthaben',
      iconUrl: 'Symbol-URL',
      iconUrlPlaceholder: 'https://example.com/icon.png',
      iconUrlHint: 'Kanal-Symbol-URL, unterstuetzt PNG, JPG-Formate',
      websiteUrl: 'Website-URL',
      websiteUrlPlaceholder: 'https://example.com',
      websiteUrlHint: 'Kanal-Website oder Auflade-Seiten-URL',
      recharge: 'Aufladen',
      visitWebsite: 'Website besuchen',
      lastChecked: 'Zuletzt geprueft'

    },

    // Redeem Codes
    redeem: {
      title: 'Redeem Code Management',
      description: 'Generate and manage redeem codes',
      generateCodes: 'Generate Codes',
      tabs: {
        redeem: 'Einloesecodes',
        promo: 'Aktionscodes'

      },
      searchCodes: 'Search codes...',
      allTypes: 'All Types',
      allStatus: 'Alle Status',
      balance: 'Guthaben',
      concurrency: 'Parallelitaet',
      subscription: 'Abonnement',
      invitation: 'Einladung',
      invitationHint: 'Einladungscodes werden verwendet, um die Benutzerregistrierung einzuschraenken. Sie werden nach Verwendung automatisch als verwendet markiert.',
      unused: 'Unbenutzt',
      used: 'Verwendet',
      columns: {
        code: 'Code',
        type: 'Typ',
        value: 'Wert',
        status: 'Status',
        usedBy: 'Verwendet von',
        usedAt: 'Verwendet am',
        createdAt: 'Erstellt am',
        actions: 'Aktionen'

      },
      form: {
        typeLabel: 'Typ',
        selectType: 'Select type',
        valueLabel: 'Wert',
        valuePlaceholder: 'Wert eingeben',
        balanceHint: 'Balance amount (USD)',
        concurrencyHint: 'Concurrency increment',
        countLabel: 'Anzahl',
        countPlaceholder: 'Anzahl eingeben',
        countHint: 'Anzahl der zu generierenden Einloesecodes',
        prefixLabel: 'Praefix (optional)',
        prefixPlaceholder: 'z.B. GIFT',
        expiresLabel: 'Laeuft ab am (optional)'

      },
      filters: {
        type: 'Typ',
        allTypes: 'All Types',
        status: 'Status',
        allStatuses: 'Alle Status',
        search: 'Search codes'
      },
      copyCode: 'Kopieren',
      disableCode: 'Deaktivieren',
      enableCode: 'Aktivieren',
      deleteConfirmMessage: 'Moechten Sie diesen Einloesecode wirklich loeschen?',
      noCodes: 'No redeem codes',
      noCodesDescription: 'Generate redeem codes to distribute balance or concurrency to users.',
      codesGeneratedSuccess: 'Generated {count} redeem code(s) successfully',
      codeDisabledSuccess: 'Redeem code disabled',
      codeEnabledSuccess: 'Redeem code enabled',
      codeDeletedSuccess: 'Redeem code deleted successfully',
      failedToUpdate: 'Failed to update redeem code',
      userPrefix: 'User #{id}',
      exportCsv: 'Export CSV',
      deleteAllUnused: 'Delete All Unused Codes',
      deleteCode: 'Delete Redeem Code',
      deleteCodeConfirm:
        'Are you sure you want to delete this redeem code? This action cannot be undone.',
      deleteAllUnusedConfirm:
        'Are you sure you want to delete all unused (active) redeem codes? This action cannot be undone.',
      deleteAll: 'Alle loeschen',
      generateCodesTitle: 'Generate Redeem Codes',
      generatedSuccessfully: 'Generated Successfully',
      codesCreated: '{count} Einloesecode(s) erstellt',
      codeType: 'Code Type',
      amount: 'Betrag ($)',
      value: 'Wert',
      count: 'Anzahl',
      generating: 'Wird generiert...',
      generate: 'Generieren',
      copyAll: 'Copy All',
      copied: 'Kopiert!',
      download: 'Herunterladen',
      codesExported: 'Codes erfolgreich exportiert',
      codeDeleted: 'Redeem code deleted successfully',
      codesDeleted: 'Successfully deleted {count} unused code(s)',
      noUnusedCodes: 'Keine unbenutzten Codes zum Loeschen',
      failedToLoad: 'Failed to load redeem codes',
      failedToGenerate: 'Failed to generate codes',
      failedToExport: 'Failed to export codes',
      failedToDelete: 'Failed to delete code',
      failedToDeleteUnused: 'Failed to delete unused codes',
      failedToCopy: 'Failed to copy codes',
      types: {
        balance: 'Guthaben',
        concurrency: 'Parallelitaet',
        subscription: 'Abonnement',
        invitation: 'Einladung',

        // Admin adjustment types (created when admin modifies user balance/concurrency)
        admin_balance: 'Balance (Admin)',
        admin_concurrency: 'Concurrency (Admin)'
      },
      selectGroup: 'Gruppe auswaehlen',
      selectGroupPlaceholder: 'Abonnement-Gruppe waehlen',
      validityDays: 'Gueltigkeitstage',
      groupRequired: 'Bitte Abonnement-Gruppe auswaehlen',
      days: ' Tage',
      status: {
        unused: 'Unbenutzt',
        used: 'Verwendet',
        expired: 'Abgelaufen',
        disabled: 'Deaktiviert'

      }
    },

    // Announcements
    announcements: {
      title: 'Ankuendigungen',
      description: 'Create announcements and target by conditions',
      createAnnouncement: 'Create Announcement',
      editAnnouncement: 'Edit Announcement',
      deleteAnnouncement: 'Delete Announcement',
      searchAnnouncements: 'Search announcements...',
      status: 'Status',
      allStatus: 'Alle Status',
      columns: {
        title: 'Titel',
        status: 'Status',
        targeting: 'Zielgruppe',
        timeRange: 'Zeitplan',
        createdAt: 'Erstellt am',
        actions: 'Aktionen'

      },
      statusLabels: {
        draft: 'Entwurf',
        active: 'Aktiv',
        archived: 'Archived'
      },
      form: {
        title: 'Titel',
        content: 'Inhalt (Markdown unterstuetzt)',
        status: 'Status',
        startsAt: 'Beginnt am',
        endsAt: 'Endet am',
        startsAtHint: 'Leer lassen fuer sofortigen Start',
        endsAtHint: 'Leer lassen fuer kein Ablaufdatum',
        targetingMode: 'Zielgruppe',
        targetingAll: 'Alle Benutzer',
        targetingCustom: 'Benutzerdefinierte Regeln',
        addOrGroup: 'ODER-Gruppe hinzufuegen',
        addAndCondition: 'UND-Bedingung hinzufuegen',
        conditionType: 'Bedingungstyp',
        conditionSubscription: 'Abonnement',
        conditionBalance: 'Guthaben',
        operator: 'Operator',
        balanceValue: 'Balance threshold',
        selectPackages: 'Pakete auswaehlen'

      },
      operators: {
        gt: '>',
        gte: '≥',
        lt: '<',
        lte: '≤',
        eq: '='
      },
      targetingSummaryAll: 'Alle Benutzer',
      targetingSummaryCustom: 'Benutzerdefiniert ({groups} Gruppen)',
      timeImmediate: 'Sofort',
      timeNever: 'Nie',
      readStatus: 'Read Status',
      eligible: 'Berechtigt',
      readAt: 'Read at',
      unread: 'Ungelesen',
      searchUsers: 'Search users...',
      failedToLoad: 'Failed to load announcements',
      failedToCreate: 'Failed to create announcement',
      failedToUpdate: 'Failed to update announcement',
      failedToDelete: 'Failed to delete announcement',
      failedToLoadReadStatus: 'Failed to load read status',
      deleteConfirm: 'Moechten Sie diese Ankuendigung wirklich loeschen? Diese Aktion kann nicht rueckgaengig gemacht werden.'

    },

    // Promo Codes
    promo: {
      title: 'Promo Code Management',
      description: 'Create and manage registration promo codes',
      createCode: 'Create Promo Code',
      editCode: 'Edit Promo Code',
      deleteCode: 'Delete Promo Code',
      searchCodes: 'Search codes...',
      allStatus: 'Alle Status',
      columns: {
        code: 'Code',
        bonusAmount: 'Bonusbetrag',
        maxUses: 'Max. Verwendungen',
        usedCount: 'Verwendet',
        usage: 'Nutzung',
        status: 'Status',
        expiresAt: 'Laeuft ab am',
        createdAt: 'Erstellt am',
        actions: 'Aktionen'

      },
      // Form labels (flat structure for template usage)
      code: 'Aktionscode',
      autoGenerate: 'automatisch generieren wenn leer',
      codePlaceholder: 'Aktionscode eingeben oder leer lassen',
      bonusAmount: 'Bonusbetrag ($)',
      maxUses: 'Max. Verwendungen',
      zeroUnlimited: '0 = unbegrenzt',
      expiresAt: 'Laeuft ab am',
      notes: 'Notizen',
      notesPlaceholder: 'Optionale Notizen fuer diesen Code',
      status: 'Status',
      neverExpires: 'Never expires',
      // Status labels
      statusActive: 'Aktiv',
      statusDisabled: 'Deaktiviert',
      statusExpired: 'Abgelaufen',
      statusMaxUsed: 'Aufgebraucht',

      // Usage records
      usageRecords: 'Usage Records',
      viewUsages: 'View Usages',
      noUsages: 'No usage records yet',
      userPrefix: 'User #{id}',
      copied: 'Kopiert!',

      // Messages
      noCodesYet: 'No promo codes yet',
      createFirstCode: 'Create your first promo code to offer registration bonuses.',
      codeCreated: 'Aktionscode erfolgreich erstellt',
      codeUpdated: 'Aktionscode erfolgreich aktualisiert',
      codeDeleted: 'Aktionscode erfolgreich geloescht',
      deleteCodeConfirm: 'Moechten Sie diesen Aktionscode wirklich loeschen? Diese Aktion kann nicht rueckgaengig gemacht werden.',
      copyRegisterLink: 'Copy register link',
      registerLinkCopied: 'Register link copied to clipboard',
      failedToLoad: 'Failed to load promo codes',
      failedToCreate: 'Failed to create promo code',
      failedToUpdate: 'Failed to update promo code',
      failedToDelete: 'Failed to delete promo code',
      failedToLoadUsages: 'Failed to load usage records'
    },

    // Usage Records
    usage: {
      title: 'Usage Records',
      description: 'Alle Benutzer-Nutzungseintraege anzeigen und verwalten',
      userFilter: 'Benutzer',
      searchUserPlaceholder: 'Search user by email...',
      searchApiKeyPlaceholder: 'Search API key by name...',
      searchAccountPlaceholder: 'Search account by name...',
      selectedUser: 'Ausgewaehlt',
      user: 'Benutzer',
      account: 'Konto',
      group: 'Gruppe',
      requestId: 'Anfrage-ID',
      requestIdCopied: 'Anfrage-ID kopiert',
      allModels: 'All Models',
      allAccounts: 'All Accounts',
      allGroups: 'All Groups',
      allTypes: 'All Types',
      inputCost: 'Input Cost',
      outputCost: 'Output Cost',
      cacheCreationCost: 'Cache Creation Cost',
      cacheReadCost: 'Cache Read Cost',
      inputTokens: 'Eingabe-Tokens',
      outputTokens: 'Ausgabe-Tokens',
      cacheCreationTokens: 'Cache-Erstellungs-Tokens',
      cacheReadTokens: 'Cache Read Tokens',
      failedToLoad: 'Failed to load usage records',
      billingType: 'Abrechnungstyp',
      allBillingTypes: 'All Billing Types',
      billingTypeBalance: 'Guthaben',
      billingTypeSubscription: 'Abonnement',
      ipAddress: 'IP',
      cleanup: {
        button: 'Bereinigung',
        title: 'Cleanup Usage Records',
        warning: 'Cleanup is irreversible and will affect historical stats.',
        submit: 'Submit Cleanup',
        submitting: 'Submitting...',
        confirmTitle: 'Confirm Cleanup',
        confirmMessage: 'Moechten Sie diese Bereinigungsaufgabe wirklich einreichen? Diese Aktion kann nicht rueckgaengig gemacht werden.',
        confirmSubmit: 'Confirm Cleanup',
        cancel: 'Abbrechen',
        cancelConfirmTitle: 'Confirm Cancel',
        cancelConfirmMessage: 'Moechten Sie diese Bereinigungsaufgabe wirklich abbrechen?',
        cancelConfirm: 'Confirm Cancel',
        cancelSuccess: 'Cleanup task canceled',
        cancelFailed: 'Failed to cancel cleanup task',
        recentTasks: 'Recent Cleanup Tasks',
        loadingTasks: 'Aufgaben werden geladen...',
        noTasks: 'Noch keine Bereinigungsaufgaben',
        range: 'Bereich',
        deletedRows: 'Geloescht',
        missingRange: 'Bitte Datumsbereich auswaehlen',
        submitSuccess: 'Cleanup task created',
        submitFailed: 'Failed to create cleanup task',
        loadFailed: 'Failed to load cleanup tasks',
        status: {
          pending: 'Ausstehend',
          running: 'Laeuft',
          succeeded: 'Succeeded',
          failed: 'Fehlgeschlagen',
          canceled: 'Canceled'
        }
      }
    },

    // Ops Monitoring
    ops: {
      title: 'Ops-Ueberwachung',
      description: 'Operational monitoring and troubleshooting',
      // Dashboard
      systemHealth: 'System Health',
      overview: 'Uebersicht',
      noSystemMetrics: 'Noch keine Systemmetriken gesammelt.',
      collectedAt: 'Gesammelt am:',
      window: 'Fenster',
      memory: 'Speicher',
      db: 'DB',
      goroutines: 'Goroutinen',
      jobs: 'Aufgaben',
      jobsHelp: 'Click “Details” to view job heartbeats and recent errors',
      active: 'active',
      idle: 'leerlauf',
      waiting: 'waiting',
      conns: 'Verb.',
      queue: 'Warteschlange',
      accountSwitches: 'Account switches',
      ok: 'ok',
      lastRun: 'last_run:',
      lastSuccess: 'last_success:',
      lastError: 'last_error:',
      noData: 'No data.',
      loadingText: 'loading',
      ready: 'bereit',
      requestsTotal: 'Requests (total)',
      slaScope: 'SLA-Bereich:',
      tokens: 'Tokens',
      tps: 'TPS:',
      current: 'current',
      peak: 'Spitze',
      average: 'average',
      totalRequests: 'Total Requests',
      avgQps: 'Avg QPS',
      avgTps: 'Avg TPS',
      avgLatency: 'Avg Request Duration',
      avgTtft: 'Durchschn. TTFT',
      exceptions: 'Exceptions',
      requestErrors: 'Request Errors',
      errorCount: 'Error Count',
      upstreamErrors: 'Upstream Errors',
      errorCountExcl429529: 'Error Count (excl 429/529)',
      sla: 'SLA (ohne Geschaeftslimits)',
      businessLimited: 'business_limited:',
      errors: 'Errors',
      errorRate: 'error_rate:',
      upstreamRate: 'upstream_rate:',
      latencyDuration: 'Request Duration',
      ttftLabel: 'TTFT (first_token_ms)',
      p50: 'p50:',
      p90: 'p90:',
      p95: 'p95:',
      p99: 'p99:',
      avg: 'avg:',
      max: 'max:',
      requests: 'Anfragen',
      requestsTitle: 'Anfragen',
      upstream: 'Upstream',
      client: 'Client',
      system: 'System',
      other: 'Andere',
      errorsSla: 'Errors (SLA scope)',
      upstreamExcl429529: 'Upstream (ohne 429/529)',
      failedToLoadData: 'Failed to load ops data.',
      failedToLoadOverview: 'Failed to load overview',
      failedToLoadThroughputTrend: 'Failed to load throughput trend',
      failedToLoadSwitchTrend: 'Failed to load avg account switches trend',
      failedToLoadLatencyHistogram: 'Failed to load request duration histogram',
      failedToLoadErrorTrend: 'Failed to load error trend',
      failedToLoadErrorDistribution: 'Failed to load error distribution',
      failedToLoadErrorDetail: 'Failed to load error detail',
      retryFailed: 'Retry failed',
      tpsK: 'TPS (K)',
      top: 'Top:',
      throughputTrend: 'Throughput Trend',
      switchRateTrend: 'Avg Account Switches',
      latencyHistogram: 'Request Duration Histogram',
      errorTrend: 'Error Trend',
      errorDistribution: 'Error Distribution',
      switchRate: 'Durchschn. Wechsel',

      // Health Score & Diagnosis
      health: 'Gesundheit',
      healthCondition: 'Health Condition',
      healthHelp: 'Gesamtsystem-Gesundheitsbewertung basierend auf SLA, Fehlerrate und Ressourcennutzung',
      healthyStatus: 'Gesund',
      riskyStatus: 'At Risk',
      idleStatus: 'Leerlauf',
      result: 'Result',
      timeRange: {
        '5m': 'Letzte 5 Minuten',
        '30m': 'Letzte 30 Minuten',
        '1h': 'Letzte 1 Stunde',
        '6h': 'Letzte 6 Stunden',
        '24h': 'Letzte 24 Stunden',
        '7d': 'Letzte 7 Tage',
        '30d': 'Letzte 30 Tage',
        custom: 'Benutzerdefiniert'

      },
      customTimeRange: {
        startTime: 'Start Time',
        endTime: 'End Time'
      },
      fullscreen: {
        enter: 'Vollbildmodus'

      },
      diagnosis: {
        title: 'Smart Diagnosis',
        footer: 'Automated diagnostic suggestions based on current metrics',
        idle: 'System is currently idle',
        idleImpact: 'No active traffic',
        // Resource diagnostics
        dbDown: 'Datenbankverbindung fehlgeschlagen',
        dbDownImpact: 'Alle Datenbankoperationen werden fehlschlagen',
        dbDownAction: 'Datenbankdienststatus, Netzwerkverbindung und Verbindungskonfiguration pruefen',
        redisDown: 'Redis-Verbindung fehlgeschlagen',
        redisDownImpact: 'Cache-Funktionalitaet eingeschraenkt, Leistung kann abnehmen',
        redisDownAction: 'Redis-Dienststatus und Netzwerkverbindung pruefen',
        cpuCritical: 'CPU-Auslastung kritisch hoch ({usage}%)',
        cpuCriticalImpact: 'System response slowing, may affect all requests',
        cpuCriticalAction: 'CPU-intensive Aufgaben pruefen, Skalierung oder Code-Optimierung erwaegen',
        cpuHigh: 'CPU-Auslastung erhoeht ({usage}%)',
        cpuHighImpact: 'Systemlast ist hoch, erfordert Aufmerksamkeit',
        cpuHighAction: 'CPU-Trends ueberwachen, Skalierungsplan vorbereiten',
        memoryCritical: 'Memory usage critically high ({usage}%)',
        memoryCriticalImpact: 'Kann OOM ausloesen, Systemstabilitaet gefaehrdet',
        memoryCriticalAction: 'Auf Speicherlecks pruefen, Speichererhoehung oder Nutzungsoptimierung erwaegen',
        memoryHigh: 'Memory usage elevated ({usage}%)',
        memoryHighImpact: 'Memory pressure is high, needs attention',
        memoryHighAction: 'Speichertrends ueberwachen, auf Speicherlecks pruefen',
        ttftHigh: 'Time to first token elevated ({ttft}ms)',
        ttftHighImpact: 'User perceived latency increased',
        ttftHighAction: 'Anfrageverarbeitungsfluss optimieren, Vorverarbeitungszeit reduzieren',

        // Error rate diagnostics
        upstreamCritical: 'Upstream-Fehlerrate kritisch hoch ({rate}%)',
        upstreamCriticalImpact: 'May affect many user requests',
        upstreamCriticalAction: 'Upstream-Dienstgesundheit pruefen, Fallback-Strategien aktivieren',
        upstreamHigh: 'Upstream-Fehlerrate erhoeht ({rate}%)',
        upstreamHighImpact: 'Empfehlung: Upstream-Dienststatus pruefen',
        upstreamHighAction: 'Upstream-Dienstteam kontaktieren, Fallback-Plan vorbereiten',
        errorHigh: 'Error rate too high ({rate}%)',
        errorHighImpact: 'Many requests failing',
        errorHighAction: 'Fehlerprotokolle pruefen, Ursache identifizieren, dringende Behebung erforderlich',
        errorElevated: 'Error rate elevated ({rate}%)',
        errorElevatedImpact: 'Empfehlung: Fehlerprotokolle pruefen',
        errorElevatedAction: 'Fehlertypen und -verteilung analysieren, Behebungsplan erstellen',

        // SLA diagnostics
        slaCritical: 'SLA kritisch unter Ziel ({sla}%)',
        slaCriticalImpact: 'User experience severely degraded',
        slaCriticalAction: 'Dringend Fehler und Latenz untersuchen, Ratenlimitierung erwaegen',
        slaLow: 'SLA unter Ziel ({sla}%)',
        slaLowImpact: 'Servicequalitaet erfordert Aufmerksamkeit',
        slaLowAction: 'SLA-Rueckgangsursachen analysieren, Systemleistung optimieren',

        // Health score diagnostics
        healthCritical: 'Gesamt-Gesundheitsbewertung kritisch niedrig ({score})',
        healthCriticalImpact: 'Mehrere Metriken koennen beeintraechtigt sein; Fehlerrate und Latenz-Untersuchung priorisieren',
        healthCriticalAction: 'Umfassende Systempruefung, kritische Probleme priorisieren',
        healthLow: 'Gesamt-Gesundheitsbewertung niedrig ({score})',
        healthLowImpact: 'Kann auf geringe Instabilitaet hinweisen; SLA und Fehlerraten ueberwachen',
        healthLowAction: 'Metrik-Trends ueberwachen, Eskalation von Problemen verhindern',
        healthy: 'Alle Systemmetriken normal',
        healthyImpact: 'Dienst laeuft stabil'

      },
      // Error Log
      errorLog: {
        timeId: 'Time / ID',
        commonErrors: {
          contextDeadlineExceeded: 'Kontext-Deadline ueberschritten',
          connectionRefused: 'Verbindung abgelehnt',
          rateLimit: 'Ratenlimit'

        },
        time: 'Zeit',
        type: 'Typ',
        context: 'Kontext',
        platform: 'Plattform',
        model: 'Modell',
        group: 'Gruppe',
        user: 'Benutzer',
        userId: 'User ID',
        account: 'Konto',
        accountId: 'Account ID',
        status: 'Status',
        message: 'Nachricht',
        latency: 'Request Duration',
        action: 'Aktion',
        noErrors: 'No errors in this window.',
        grp: 'GRP:',
        acc: 'ACC:',
        details: 'Details',
        phase: 'Phase',
        id: 'ID:',
        typeUpstream: 'Upstream',
        typeRequest: 'Anfrage',
        typeAuth: 'Auth',
        typeRouting: 'Routing',
        typeInternal: 'Intern'

      },
      // Error Details Modal
      errorDetails: {
        upstreamErrors: 'Upstream Errors',
        requestErrors: 'Request Errors',
        unresolved: 'Ungeloest',
        resolved: 'Geloest',
        viewErrors: 'Errors',
        viewExcluded: 'Ausgeschlossen',
        statusCodeOther: 'Andere',
        owner: {
          provider: 'Anbieter',
          client: 'Client',
          platform: 'Plattform'

        },
        phase: {
          request: 'Anfrage',
          auth: 'Auth',
          routing: 'Routing',
          upstream: 'Upstream',
          network: 'Netzwerk',
          internal: 'Intern'

        },
        total: 'Total:',
        searchPlaceholder: 'Search request_id / client_request_id / message',
      },
      // Error Detail Modal
      errorDetail: {
        title: 'Error Detail',
        titleWithId: 'Error #{id}',
        noErrorSelected: 'Kein Fehler ausgewaehlt.',
        resolution: 'Geloest:',
        pinnedToOriginalAccountId: 'An Original-account_id angeheftet',
        missingUpstreamRequestBody: 'Fehlender Upstream-Anfrage-Body',
        failedToLoadRetryHistory: 'Failed to load retry history',
        failedToUpdateResolvedStatus: 'Failed to update resolved status',
        unsupportedRetryMode: 'Nicht unterstuetzter Wiederholungsmodus',
        classificationKeys: {
          phase: 'Phase',
          owner: 'Eigentuemer',
          source: 'Quelle',
          retryable: 'Retryable',
          resolvedAt: 'Geloest am',
          resolvedBy: 'Geloest von',
          resolvedRetryId: 'Resolved Retry',
          retryCount: 'Retry Count'
        },
        source: {
          upstream_http: 'Upstream HTTP'

        },
        upstreamKeys: {
          status: 'Status',
          message: 'Nachricht',
          detail: 'Detail',
          upstreamErrors: 'Upstream Errors'
        },
        upstreamEvent: {
          account: 'Konto',
          status: 'Status',
          requestId: 'Anfrage-ID'

        },
        responsePreview: {
          expand: 'Antwort (klicken zum Aufklappen)',
          collapse: 'Antwort (klicken zum Zuklappen)'

        },
        retryMeta: {
          used: 'Verwendet',
          success: 'Erfolg',
          pinned: 'Angeheftet'

        },
        loading: 'Wird geladen…',
        requestId: 'Anfrage-ID',
        time: 'Zeit',
        phase: 'Phase',
        status: 'Status',
        message: 'Nachricht',
        basicInfo: 'Grundinformationen',
        platform: 'Plattform',
        model: 'Modell',
        group: 'Gruppe',
        user: 'Benutzer',
        account: 'Konto',
        latency: 'Request Duration',
        businessLimited: 'Business Limited',
        requestPath: 'Anfragepfad',
        timings: 'Zeitmessungen',
        auth: 'Auth',
        routing: 'Routing',
        upstream: 'Upstream',
        response: 'Antwort',
        classification: 'Klassifizierung',
        notRetryable: 'Wiederholung nicht empfohlen',
        retry: 'Wiederholen',
        retryClient: 'Retry (Client)',
        retryUpstream: 'Retry (Upstream pinned)',
        pinnedAccountId: 'Angeheftete account_id',
        retryNotes: 'Retry Notes',
        requestBody: 'Anfrage-Body',
        errorBody: 'Error Body',
        trimmed: 'gekuerzt',
        confirmRetry: 'Confirm Retry',
        retrySuccess: 'Retry succeeded',
        retryFailed: 'Retry failed',
        retryHint: 'Retry will resend the request with the same parameters',
        retryClientHint: 'Client-Wiederholung verwenden (kein Konto-Anheften)',
        retryUpstreamHint: 'Upstream-angeheftete Wiederholung verwenden (an das Fehlerkonto anheften)',
        pinnedAccountIdHint: '(automatisch aus Fehlerprotokoll)',
        retryNote1: 'Retry will use the same request body and parameters',
        retryNote2: 'Wenn die urspruengliche Anfrage aufgrund von Kontoproblemen fehlschlug, kann die angeheftete Wiederholung immer noch fehlschlagen',
        retryNote3: 'Client-Wiederholung waehlt ein neues Konto',
        retryNote4: 'Sie koennen eine erzwungene Wiederholung fuer nicht wiederholbare Fehler durchfuehren, aber es wird nicht empfohlen',
        confirmRetryMessage: 'Confirm retry this request?',
        confirmRetryHint: 'Wird mit denselben Anfrageparametern erneut gesendet',
        forceRetry: 'Ich verstehe und moechte die Wiederholung erzwingen',
        forceRetryHint: 'Dieser Fehler kann normalerweise nicht durch Wiederholung behoben werden; aktivieren Sie das Kontrollkaestchen, um fortzufahren',
        forceRetryNeedAck: 'Bitte aktivieren, um Wiederholung zu erzwingen',
        markResolved: 'Als geloest markieren',
        markUnresolved: 'Als ungeloest markieren',
        viewRetries: 'Retry history',
        retryHistory: 'Retry History',
        tabOverview: 'Uebersicht',
        tabRetries: 'Wiederholungen',
        tabRequest: 'Anfrage',
        tabResponse: 'Antwort',
        responseBody: 'Antwort',
        compareA: 'Vergleich A',
        compareB: 'Vergleich B',
        retrySummary: 'Retry Summary',
        responseHintSucceeded: 'Showing succeeded retry response_preview (#{id})',
        responseHintFallback: 'Keine erfolgreiche Wiederholung gefunden; gespeicherter error_body wird angezeigt',
        suggestion: 'Vorschlag',
        suggestUpstreamResolved: '✓ Upstream error resolved by retry; no action needed',
        suggestUpstream: 'Upstream-Instabilitaet: Kontostatus pruefen, Kontowechsel erwaegen oder wiederholen',
        suggestRequest: 'Client-Anfragefehler: Kunden bitten, Anfrageparameter zu korrigieren',
        suggestAuth: 'Authentifizierung fehlgeschlagen: API-Schluessel/Zugangsdaten ueberpruefen',
        suggestPlatform: 'Plattformfehler: Untersuchung und Behebung priorisieren',
        suggestGeneric: 'Weitere Details fuer mehr Kontext anzeigen'

      },
      requestDetails: {
        title: 'Anfragedetails',
        details: 'Details',
        rangeLabel: 'Fenster: {range}',
        rangeMinutes: '{n} Minuten',
        rangeHours: '{n} Stunden',
        empty: 'No requests in this window.',
        emptyHint: 'Versuchen Sie einen anderen Zeitraum oder entfernen Sie Filter.',
        failedToLoad: 'Failed to load request details',
        requestIdCopied: 'Anfrage-ID kopiert',
        copyFailed: 'Copy failed',
        copy: 'Kopieren',
        viewError: 'View Error',
        kind: {
          success: 'ERFOLGREICH',
          error: 'FEHLER'

        },
        table: {
          time: 'Zeit',
          kind: 'Art',
          platform: 'Plattform',
          model: 'Modell',
          duration: 'Duration',
          status: 'Status',
          requestId: 'Anfrage-ID',
          actions: 'Aktionen'

        }
      },
      alertEvents: {
        title: 'Alarmereignisse',
        description: 'Aktuelle Alarm-Ausloesung/Aufloesung-Eintraege (nur E-Mail)',
        loading: 'Wird geladen...',
        empty: 'Keine Alarmereignisse',
        loadFailed: 'Failed to load alert events',
        status: {
          firing: 'AKTIV',
          resolved: 'GELOEST',
          manualResolved: 'MANUELL GELOEST'

        },
        detail: {
          title: 'Alarmdetail',
          loading: 'Details werden geladen...',
          empty: 'Kein Detail',
          loadFailed: 'Failed to load alert detail',
          manualResolve: 'Als geloest markieren',
          manualResolvedSuccess: 'Als manuell geloest markiert',
          manualResolvedFailed: 'Failed to mark as manually resolved',
          silence: 'Alarm ignorieren',
          silenceSuccess: 'Alarm stummgeschaltet',
          silenceFailed: 'Failed to silence alert',
          viewRule: 'Regel anzeigen',
          viewLogs: 'Protokolle anzeigen',
          firedAt: 'Ausgeloest am',
          resolvedAt: 'Geloest am',
          ruleId: 'Regel-ID',
          dimensions: 'Dimensionen',
          historyTitle: 'Verlauf',
          historyHint: 'Aktuelle Ereignisse mit gleicher Regel + Dimensionen',
          historyLoading: 'Verlauf wird geladen...',
          historyEmpty: 'Kein Verlauf'

        },
        table: {
          time: 'Zeit',
          status: 'Status',
          severity: 'Schweregrad',
          platform: 'Plattform',
          ruleId: 'Regel-ID',
          title: 'Titel',
          duration: 'Duration',
          metric: 'Metrik / Schwellenwert',
          dimensions: 'Dimensionen',
          email: 'Email Sent',
          emailSent: 'Gesendet',
          emailIgnored: 'Ignoriert'

        }
      },
      alertRules: {
        title: 'Alarmregeln',
        description: 'Create and manage threshold-based system alerts (email-only)',
        loading: 'Wird geladen...',
        empty: 'Keine Alarmregeln',
        loadFailed: 'Failed to load alert rules',
        saveFailed: 'Failed to save alert rule',
        saveSuccess: 'Alarmregel erfolgreich gespeichert',
        deleteFailed: 'Failed to delete alert rule',
        deleteSuccess: 'Alarmregel erfolgreich geloescht',
        manage: 'Manage Alert Rules',
        create: 'Create Rule',
        createTitle: 'Create Alert Rule',
        editTitle: 'Edit Alert Rule',
        deleteConfirmTitle: 'Delete this rule?',
        deleteConfirmMessage: 'This will remove the rule and its related events. Continue?',
        metricGroups: {
          system: 'Systemmetriken',
          group: 'Gruppenebene-Metriken (erfordert group_id)',
          account: 'Account-level Metrics'
        },
        metrics: {
          successRate: 'Success Rate (%)',
          errorRate: 'Error Rate (%)',
          upstreamErrorRate: 'Upstream Error Rate (%)',
          p95: 'P95-Latenz (ms)',
          p99: 'P99-Latenz (ms)',
          cpu: 'CPU Usage (%)',
          memory: 'Memory Usage (%)',
          queueDepth: 'Concurrency Queue Depth',
          groupAvailableAccounts: 'Group Available Accounts',
          groupAvailableRatio: 'Group Available Ratio (%)',
          groupRateLimitRatio: 'Group Rate Limit Ratio (%)',
          accountRateLimitedCount: 'Rate-limited Accounts',
          accountErrorCount: 'Error Accounts (excluding temporarily unschedulable)',
          accountErrorRatio: 'Error Account Ratio (%)',
          overloadAccountCount: 'Overloaded Accounts'
        },
        metricDescriptions: {
          successRate: 'Percentage of successful requests in the window (0-100).',
          errorRate: 'Percentage of failed requests in the window (0-100).',
          upstreamErrorRate: 'Prozentsatz der Upstream-Fehler im Fenster (0-100).',
          p95: 'P95-Anfragelatenz innerhalb des Fensters (ms).',
          p99: 'P99-Anfragelatenz innerhalb des Fensters (ms).',
          cpu: 'Aktuelle Instanz-CPU-Auslastung (0-100).',
          memory: 'Aktuelle Instanz-Speicherauslastung (0-100).',
          queueDepth: 'Concurrency queue depth within the window (queued requests).',
          groupAvailableAccounts: 'Anzahl verfuegbarer Konten in der ausgewaehlten Gruppe (erfordert group_id).',
          groupAvailableRatio: 'Available account ratio in the selected group (0-100, requires group_id).',
          groupRateLimitRatio: 'Rate-limited account ratio in the selected group (0-100, requires group_id).',
          accountRateLimitedCount: 'Anzahl ratenbegrenzter Konten innerhalb des Fensters.',
          accountErrorCount: 'Anzahl fehlerhafter Konten innerhalb des Fensters (ohne temporaer nichtplanbare).',
          accountErrorRatio: 'Error account ratio within the window (0-100).',
          overloadAccountCount: 'Anzahl ueberlasteter Konten innerhalb des Fensters.'

        },
        hints: {
          recommended: 'Recommended: operator {operator}, threshold {threshold}{unit}',
          groupRequired: 'Dies ist eine Gruppenebene-Metrik; die Auswahl einer Gruppe (group_id) ist erforderlich.',
          groupOptional: 'Optional: Regel auf eine bestimmte Gruppe ueber group_id beschraenken.'

        },
        table: {
          name: 'Name',
          metric: 'Metrik',
          severity: 'Schweregrad',
          enabled: 'Aktiviert',
          actions: 'Aktionen'

        },
        form: {
          name: 'Name',
          description: 'Beschreibung',
          metric: 'Metrik',
          operator: 'Operator',
          groupId: 'Gruppe (group_id)',
          groupPlaceholder: 'Select a group',
          allGroups: 'Alle Gruppen',
          threshold: 'Schwellenwert',
          severity: 'Schweregrad',
          window: 'Fenster (Minuten)',
          sustained: 'Anhalten (Stichproben)',
          cooldown: 'Cooldown (minutes)',
          enabled: 'Aktiviert',
          notifyEmail: 'E-Mail-Benachrichtigungen senden'

        },
        validation: {
          title: 'Bitte beheben Sie die folgenden Probleme',
          invalid: 'Ungueltige Regel',
          nameRequired: 'Name ist erforderlich',
          metricRequired: 'Metrik ist erforderlich',
          groupIdRequired: 'group_id ist fuer Gruppenebene-Metriken erforderlich',
          operatorRequired: 'Operator ist erforderlich',
          thresholdRequired: 'Schwellenwert muss eine Zahl sein',
          windowRange: 'Fenster muss eines sein von: 1, 5, 60 Minuten',
          sustainedRange: 'Anhalten muss zwischen 1 und 1440 Stichproben liegen',
          cooldownRange: 'Cooldown must be between 0 and 1440 minutes'
        }
      },
      runtime: {
        title: 'Ops Runtime Settings',
        description: 'In Datenbank gespeichert; Aenderungen werden wirksam ohne Konfigurationsdateien zu bearbeiten.',
        loading: 'Wird geladen...',
        noData: 'Keine Laufzeiteinstellungen verfuegbar',
        loadFailed: 'Failed to load runtime settings',
        saveSuccess: 'Laufzeiteinstellungen gespeichert',
        saveFailed: 'Failed to save runtime settings',
        alertTitle: 'Alarm-Evaluator',
        groupAvailabilityTitle: 'Gruppenverfuegbarkeits-Monitor',
        evalIntervalSeconds: 'Evaluationsintervall (Sekunden)',
        silencing: {
          title: 'Alarm-Stummschaltung (Wartungsmodus)',
          enabled: 'Enable silencing',
          globalUntil: 'Stummschalten bis (RFC3339)',
          untilHint: 'Leer lassen, um die Stummschaltung nur umzuschalten, ohne Ablauf (nicht empfohlen).',
          reason: 'Grund',
          reasonPlaceholder: 'z.B. geplante Wartung',
          entries: {
            title: 'Advanced: targeted silencing',
            hint: 'Optional: nur bestimmte Regeln oder Schweregrade stummschalten. Felder leer lassen, um alle abzugleichen.',
            add: 'Eintrag hinzufuegen',
            empty: 'Keine gezielten Eintraege',
            entryTitle: 'Eintrag #{n}',
            ruleId: 'Regel-ID (optional)',
            ruleIdPlaceholder: 'z.B. 1',
            severities: 'Schweregrade (optional)',
            severitiesPlaceholder: 'z.B. P0,P1 (leer = alle)',
            until: 'Bis (RFC3339)',
            reason: 'Grund',
            validation: {
              untilRequired: 'Eintrags-Endzeit ist erforderlich',
              untilFormat: 'Eintrags-Endzeit muss ein gueltiger RFC3339-Zeitstempel sein',
              ruleIdPositive: 'Eintrags-rule_id muss eine positive Ganzzahl sein',
              severitiesFormat: 'Eintrags-Schweregrade muessen eine kommagetrennte Liste von P0..P3 sein'

            }
          },
          validation: {
            timeFormat: 'Stummschaltzeit muss ein gueltiger RFC3339-Zeitstempel sein'

          }
        },
        lockEnabled: 'Distributed Lock Enabled',
        lockKey: 'Verteilter Sperr-Schluessel',
        lockTTLSeconds: 'Verteilte Sperr-TTL (Sekunden)',
        showAdvancedDeveloperSettings: 'Erweiterte Entwicklereinstellungen anzeigen (Verteilte Sperre)',
        advancedSettingsSummary: 'Advanced settings (Distributed Lock)',
        evalIntervalHint: 'Wie oft der Evaluator laeuft. Es wird empfohlen, den Standard beizubehalten.',
        validation: {
          title: 'Bitte beheben Sie die folgenden Probleme',
          invalid: 'Ungueltige Einstellungen',
          evalIntervalRange: 'Evaluationsintervall muss zwischen 1 und 86400 Sekunden liegen',
          lockKeyRequired: 'Verteilter Sperr-Schluessel ist erforderlich, wenn die Sperre aktiviert ist',
          lockKeyPrefix: 'Verteilter Sperr-Schluessel muss mit "{prefix}" beginnen',
          lockKeyHint: 'Recommended: start with "{prefix}" to avoid conflicts',
          lockTtlRange: 'Verteilte Sperr-TTL muss zwischen 1 und 86400 Sekunden liegen',
          slaMinPercentRange: 'SLA-Mindestprozentsatz muss zwischen 0 und 100 liegen',
          ttftP99MaxRange: 'TTFT P99-Maximum muss eine Zahl ≥ 0 sein',
          requestErrorRateMaxRange: 'Maximum der Anfragefehlerrate muss zwischen 0 und 100 liegen',
          upstreamErrorRateMaxRange: 'Maximum der Upstream-Fehlerrate muss zwischen 0 und 100 liegen'

        }
      },
      email: {
        title: 'Email Notification',
        description: 'Alarm-/Bericht-E-Mail-Benachrichtigungen konfigurieren (in Datenbank gespeichert).',
        loading: 'Wird geladen...',
        noData: 'Keine E-Mail-Benachrichtigungskonfiguration',
        loadFailed: 'Failed to load email notification config',
        saveSuccess: 'Email notification config saved',
        saveFailed: 'Failed to save email notification config',
        alertTitle: 'Alert Emails',
        reportTitle: 'Report Emails',
        recipients: 'Empfaenger',
        recipientsHint: 'Wenn leer, kann das System auf die erste Admin-E-Mail zurueckfallen.',
        minSeverity: 'Min. Schweregrad',
        minSeverityAll: 'Alle Schweregrade',
        rateLimitPerHour: 'Rate limit per hour',
        batchWindowSeconds: 'Batchfenster (Sekunden)',
        includeResolved: 'Geloeste Alarme einschliessen',
        dailySummary: 'Daily summary',
        weeklySummary: 'Weekly summary',
        errorDigest: 'Error digest',
        errorDigestMinCount: 'Min. Fehler fuer Zusammenfassung',
        accountHealth: 'Account health',
        accountHealthThreshold: 'Error rate threshold (%)',
        cronPlaceholder: 'Cron-Ausdruck',
        reportHint: 'Zeitplaene verwenden Cron-Syntax; leer lassen fuer Standardwerte.',
        validation: {
          title: 'Bitte beheben Sie die folgenden Probleme',
          invalid: 'Ungueltige E-Mail-Benachrichtigungskonfiguration',
          alertRecipientsRequired: 'Alarm-E-Mails sind aktiviert, aber keine Empfaenger konfiguriert',
          reportRecipientsRequired: 'Berichts-E-Mails sind aktiviert, aber keine Empfaenger konfiguriert',
          invalidRecipients: 'Eine oder mehrere Empfaenger-E-Mails sind ungueltig',
          rateLimitRange: 'Rate limit per hour must be a number ≥ 0',
          batchWindowRange: 'Batchfenster muss zwischen 0 und 86400 Sekunden liegen',
          cronRequired: 'Ein Cron-Ausdruck ist erforderlich, wenn der Zeitplan aktiviert ist',
          cronFormat: 'Cron-Ausdrucksformat scheint ungueltig (mindestens 5 Teile erwartet)',
          digestMinCountRange: 'Min. Fehler fuer Zusammenfassung muss eine Zahl ≥ 0 sein',
          accountHealthThresholdRange: 'Account health threshold must be between 0 and 100'
        }
      },
      settings: {
        title: 'Ops Monitoring Settings',
        loadFailed: 'Failed to load settings',
        saveSuccess: 'Ops-Ueberwachungseinstellungen erfolgreich gespeichert',
        saveFailed: 'Failed to save settings',
        dataCollection: 'Datenerfassung',
        evaluationInterval: 'Evaluationsintervall (Sekunden)',
        evaluationIntervalHint: 'Haeufigkeit der Erkennungsaufgaben, Standard beibehalten empfohlen',
        alertConfig: 'Alarmkonfiguration',
        enableAlert: 'Enable Alerts',
        alertRecipients: 'Alert Recipient Emails',
        emailPlaceholder: 'E-Mail-Adresse eingeben',
        recipientsHint: 'Wenn leer, verwendet das System die erste Admin-E-Mail als Standardempfaenger',
        minSeverity: 'Mindest-Schweregrad',
        reportConfig: 'Berichtskonfiguration',
        enableReport: 'Enable Reports',
        reportRecipients: 'Report Recipient Emails',
        dailySummary: 'Daily Summary',
        weeklySummary: 'Weekly Summary',
        metricThresholds: 'Metrische Schwellenwerte',
        metricThresholdsHint: 'Alarmschwellenwerte fuer Metriken konfigurieren, Werte ueber dem Schwellenwert werden rot angezeigt',
        slaMinPercent: 'SLA-Mindestprozentsatz',
        slaMinPercentHint: 'SLA unter diesem Wert wird rot angezeigt (Standard: 99,5%)',
        ttftP99MaxMs: 'TTFT P99-Maximum (ms)',
        ttftP99MaxMsHint: 'TTFT P99 ueber diesem Wert wird rot angezeigt (Standard: 500ms)',
        requestErrorRateMaxPercent: 'Request Error Rate Maximum (%)',
        requestErrorRateMaxPercentHint: 'Anfragefehlerrate ueber diesem Wert wird rot angezeigt (Standard: 5%)',
        upstreamErrorRateMaxPercent: 'Upstream Error Rate Maximum (%)',
        upstreamErrorRateMaxPercentHint: 'Upstream-Fehlerrate ueber diesem Wert wird rot angezeigt (Standard: 5%)',
        advancedSettings: 'Advanced Settings',
        dataRetention: 'Datenaufbewahrungsrichtlinie',
        enableCleanup: 'Enable Data Cleanup',
        cleanupSchedule: 'Cleanup Schedule (Cron)',
        cleanupScheduleHint: 'Beispiel: 0 2 * * * bedeutet taeglich um 2 Uhr',
        errorLogRetentionDays: 'Error Log Retention Days',
        minuteMetricsRetentionDays: 'Minutenmetriken-Aufbewahrungstage',
        hourlyMetricsRetentionDays: 'Hourly Metrics Retention Days',
        retentionDaysHint: 'Recommended 7-90 days, longer periods will consume more storage',
        aggregation: 'Voraggregationsaufgaben',
        enableAggregation: 'Enable Pre-aggregation',
        aggregationHint: 'Voraggregation verbessert die Abfrageleistung fuer lange Zeitfenster',
        errorFiltering: 'Error Filtering',
        ignoreCountTokensErrors: 'count_tokens-Fehler ignorieren',
        ignoreCountTokensErrorsHint: 'When enabled, errors from count_tokens requests will not be written to the error log.',
        ignoreContextCanceled: 'Client-Trennungsfehler ignorieren',
        ignoreContextCanceledHint: 'Wenn aktiviert, werden Client-Trennungsfehler (context canceled) nicht ins Fehlerprotokoll geschrieben.',
        ignoreNoAvailableAccounts: '"Keine verfuegbaren Konten"-Fehler ignorieren',
        ignoreNoAvailableAccountsHint: 'Wenn aktiviert, werden "Keine verfuegbaren Konten"-Fehler nicht ins Fehlerprotokoll geschrieben (nicht empfohlen; normalerweise ein Konfigurationsproblem).',
        ignoreInvalidApiKeyErrors: 'Ungueltige API-Schluessel-Fehler ignorieren',
        ignoreInvalidApiKeyErrorsHint: 'Wenn aktiviert, werden ungueltige oder fehlende API-Schluessel-Fehler (INVALID_API_KEY, API_KEY_REQUIRED) nicht ins Fehlerprotokoll geschrieben.',
        autoRefresh: 'Auto Refresh',
        enableAutoRefresh: 'Enable auto refresh',
        enableAutoRefreshHint: 'Dashboard-Daten automatisch in festen Intervallen aktualisieren.',
        refreshInterval: 'Refresh Interval',
        refreshInterval15s: '15 Sekunden',
        refreshInterval30s: '30 Sekunden',
        refreshInterval60s: '60 Sekunden',
        autoRefreshCountdown: 'Automatische Aktualisierung: {seconds}s',
        validation: {
          title: 'Bitte beheben Sie die folgenden Probleme',
          retentionDaysRange: 'Aufbewahrungstage muessen zwischen 1-365 Tagen liegen',
          slaMinPercentRange: 'SLA-Mindestprozentsatz muss zwischen 0 und 100 liegen',
          ttftP99MaxRange: 'TTFT P99-Maximum muss eine Zahl ≥ 0 sein',
          requestErrorRateMaxRange: 'Maximum der Anfragefehlerrate muss zwischen 0 und 100 liegen',
          upstreamErrorRateMaxRange: 'Maximum der Upstream-Fehlerrate muss zwischen 0 und 100 liegen'

        }
      },
      concurrency: {
        title: 'Concurrency / Queue',
        byPlatform: 'Nach Plattform',
        byGroup: 'Nach Gruppe',
        byAccount: 'By Account',
        totalRows: '{count} Zeilen',
        disabledHint: 'Realtime monitoring is disabled in settings.',
        empty: 'Keine Daten',
        queued: 'Warteschlange {count}',
        rateLimited: 'Rate-limited {count}',
        scopeRateLimitedTooltip: '{scope} ratenbegrenzt ({count} Konten)',
        errorAccounts: 'Errors {count}',
        loadFailed: 'Failed to load concurrency data',
        switchToUser: 'Nach Benutzer wechseln',
        switchToPlatform: 'Nach Plattform wechseln'
      },
      realtime: {
        title: 'Realtime',
        connected: 'Realtime connected',
        connecting: 'Realtime connecting',
        reconnecting: 'Realtime reconnecting',
        offline: 'Realtime offline',
        closed: 'Realtime closed',
        reconnectIn: 'Wiederholung in {seconds}s'

      },
      queryMode: {
        auto: 'Auto',
        raw: 'Roh',
        preagg: 'Voragg'

      },
      accountAvailability: {
        available: 'Verfuegbar',
        unavailable: 'Nicht verfuegbar',
        accountError: 'Fehler'

      },
      tooltips: {
        totalRequests: 'Total number of requests (including both successful and failed requests) in the selected time window.',
        throughputTrend: 'Requests/QPS + Tokens/TPS in the selected window.',
        switchRateTrend: 'Trend of account switches / total requests over the last 5 hours (avg switches).',
        latencyHistogram: 'Request duration distribution (ms) for successful requests.',
        errorTrend: 'Error counts over time (SLA scope excludes business limits; upstream excludes 429/529).',
        errorDistribution: 'Error distribution by status code.',
        goroutines:
          'Number of Go runtime goroutines (lightweight threads). There is no absolute "safe" number—use your historical baseline. Heuristic: <2k is common; 2k–8k watch; >8k plus rising queue/latency often suggests blocking/leaks.',
        cpu: 'CPU-Auslastung in Prozent, zeigt die Systemprozessorlast.',
        memory: 'Memory usage, including used and total available memory.',
        db: 'Database connection pool status, including active, idle, and waiting connections.',
        redis: 'Redis connection pool status, showing active and idle connections.',
        jobs: 'Background job execution status, including last run time, success time, and error information.',
        qps: 'Anfragen pro Sekunde (QPS) und Tokens pro Sekunde (TPS), Echtzeit-Systemdurchsatz.',
        tokens: 'Total number of tokens processed in the current time window.',
        sla: 'Service Level Agreement Erfolgsrate, ohne Geschaeftslimits (z.B. unzureichendes Guthaben, Kontingent ueberschritten).',
        errors: 'Error statistics, including total errors, error rate, and upstream error rate.',
        upstreamErrors: 'Upstream-Fehlerstatistiken, ohne Ratenlimit-Fehler (429/529).',
        latency: 'Anfragedauer-Statistiken, einschliesslich p50, p90, p95, p99 Perzentile.',
        ttft: 'Time To First Token, measuring the speed of first token return in streaming responses.',
        health: 'System-Gesundheitsbewertung (0-100), unter Beruecksichtigung von SLA, Fehlerrate und Ressourcennutzung.'

      },
      charts: {
        emptyRequest: 'No requests in this window.',
        emptyError: 'No errors in this window.',
        resetZoom: 'Zuruecksetzen',
        resetZoomHint: 'Reset zoom (if enabled)',
        downloadChart: 'Herunterladen',
        downloadChartHint: 'Download chart as image'
      }
    },

    // Settings
    settings: {
      title: 'System Settings',
      description: 'Manage registration, email verification, default values, and SMTP settings',
      registration: {
        title: 'Registration Settings',
        description: 'Benutzerregistrierung und Verifizierung steuern',
        enableRegistration: 'Enable Registration',
        enableRegistrationHint: 'Neuen Benutzern die Registrierung erlauben',
        emailVerification: 'Email Verification',
        emailVerificationHint: 'E-Mail-Verifizierung fuer neue Registrierungen erfordern',
        promoCode: 'Aktionscode',
        promoCodeHint: 'Benutzern erlauben, Aktionscodes bei der Registrierung zu verwenden',
        invitationCode: 'Invitation Code Registration',
        invitationCodeHint: 'Wenn aktiviert, muessen Benutzer einen gueltigen Einladungscode zur Registrierung eingeben',
        passwordReset: 'Password Reset',
        passwordResetHint: 'Benutzern erlauben, ihr Passwort per E-Mail zurueckzusetzen',
        totp: 'Zwei-Faktor-Authentifizierung (2FA)',
        totpHint: 'Benutzern erlauben, Authenticator-Apps wie Google Authenticator zu verwenden',
        totpKeyNotConfigured:
          'Please configure TOTP_ENCRYPTION_KEY in environment variables first. Generate a key with: openssl rand -hex 32'
      },
      turnstile: {
        title: 'Cloudflare Turnstile',
        description: 'Bot-Schutz fuer Anmeldung und Registrierung',
        enableTurnstile: 'Enable Turnstile',
        enableTurnstileHint: 'Cloudflare Turnstile-Verifizierung erfordern',
        siteKey: 'Seitenschluessel',
        secretKey: 'Geheimschluessel',
        siteKeyHint: 'Erhalten Sie dies aus Ihrem Cloudflare Dashboard',
        cloudflareDashboard: 'Cloudflare Dashboard',
        secretKeyHint: 'Serverseitiger Verifizierungsschluessel (geheim halten)',
        secretKeyConfiguredHint: 'Secret key configured. Leave empty to keep the current value.'      },
      linuxdo: {
        title: 'LinuxDo Connect Login',
        description: 'LinuxDo Connect OAuth fuer Sub2API-Endbenutzeranmeldung konfigurieren',
        enable: 'Enable LinuxDo Login',
        enableHint: 'LinuxDo-Anmeldung auf den Anmelde-/Registrierungsseiten anzeigen',
        clientId: 'Client-ID',
        clientIdPlaceholder: 'e.g., hprJ5pC3...',
        clientIdHint: 'Get this from Connect.Linux.Do',
        clientSecret: 'Client Secret',
        clientSecretPlaceholder: '********',
        clientSecretHint: 'Vom Backend zum Token-Austausch verwendet (geheim halten)',
        clientSecretConfiguredPlaceholder: '********',
        clientSecretConfiguredHint: 'Secret configured. Leave empty to keep the current value.',
        redirectUrl: 'Weiterleitungs-URL',
        redirectUrlPlaceholder: 'https://your-domain.com/api/v1/auth/oauth/linuxdo/callback',
        redirectUrlHint:
          'Must match the redirect URL configured in Connect.Linux.Do (must be an absolute http(s) URL)',
        quickSetCopy: 'Generate & Copy (current site)',
        redirectUrlSetAndCopied: 'Weiterleitungs-URL generiert und in Zwischenablage kopiert'

      },
      defaults: {
        title: 'Default User Settings',
        description: 'Standardwerte fuer neue Benutzer',
        defaultBalance: 'Default Balance',
        defaultBalanceHint: 'Anfangsguthaben fuer neue Benutzer',
        defaultConcurrency: 'Default Concurrency',
        defaultConcurrencyHint: 'Maximum concurrent requests for new users'
      },
      site: {
        title: 'Site Settings',
        description: 'Website-Branding anpassen',
        siteName: 'Site Name',
        siteNamePlaceholder: 'Sub2API',
        siteNameHint: 'In E-Mails und Seitentiteln angezeigt',
        siteSubtitle: 'Untertitel der Website',
        siteSubtitlePlaceholder: 'Subscription to API Conversion Platform',
        siteSubtitleHint: 'Auf Anmelde- und Registrierungsseiten angezeigt',
        defaultLocale: 'Standardsprache',
        defaultLocaleBrowser: 'Browser folgen',
        defaultLocaleHint: 'Benutzer, die keine Sprache manuell ausgewählt haben, sehen standardmäßig diese Sprache. Leer lassen, um der Browsersprache zu folgen.',
        apiBaseUrl: 'API Basis-URL',
        apiBaseUrlPlaceholder: 'https://api.example.com',
        apiBaseUrlHint:
          'Used for "Use Key" and "Import to CC Switch" features. Leave empty to use current site URL.',
        contactInfo: 'Kontaktinformationen',
        contactInfoPlaceholder: 'e.g., QQ: 123456789',
        contactInfoHint: 'Kundensupport-Kontaktinformationen, angezeigt auf der Einloeseseite, im Profil usw.',
        docUrl: 'Documentation URL',
        docUrlPlaceholder: 'https://docs.example.com',
        docUrlHint: 'Link zu Ihrer Dokumentationsseite. Leer lassen, um den Dokumentationslink auszublenden.',
        queryDomain: 'Abfrage-Domain',
        queryDomainHint: 'Nach der Konfiguration zeigt diese Domain die API Key Abfrageseite an',
        siteLogo: 'Site Logo',
        uploadImage: 'Upload Image',
        remove: 'Remove',
        logoHint: 'PNG, JPG, or SVG. Max 300KB. Recommended: 80x80px square image.',
        logoSizeError: 'Bildgroesse ueberschreitet das 300KB-Limit ({size}KB)',
        logoTypeError: 'Bitte waehlen Sie eine Bilddatei',
        logoReadError: 'Failed to read the image file',
        homeContent: 'Home Page Content',
        homeContentPlaceholder: 'Enter custom content for the home page. Supports Markdown & HTML. If a URL is entered, it will be displayed as an iframe.',
        homeContentHint: 'Inhalt der Startseite anpassen. Unterstuetzt Markdown/HTML. Wenn Sie eine URL eingeben (beginnend mit http:// oder https://), wird sie als iframe-src verwendet, um eine externe Seite einzubetten. Wenn gesetzt, werden die Standard-Statusinformationen nicht mehr angezeigt.',
        homeContentIframeWarning: '⚠️ iframe mode note: Some websites have X-Frame-Options or CSP security policies that prevent embedding in iframes. If the page appears blank or shows an error, please verify the target website allows embedding, or consider using HTML mode to build your own content.',
        hideCcsImportButton: 'Hide CCS Import Button',
        hideCcsImportButtonHint: 'When enabled, the "Import CCS" button will be hidden on the API Keys page',
        cryptoAddresses: 'Kryptowährungsadressen',
        cryptoAddressesHint: 'Konfigurieren Sie Empfangsadressen für Kryptowährungen, die auf der Tarifseite angezeigt werden',
        addCryptoAddress: 'Adresse hinzufügen',
        chainName: 'Netzwerk',
        walletAddress: 'Adresse'
      },
      purchase: {
        title: 'Purchase Page',
        description: 'Show a "Purchase Subscription" entry in the sidebar and open the configured URL in an iframe',
        enabled: 'Show Purchase Entry',
        enabledHint: 'Nur im Standardmodus angezeigt (nicht im einfachen Modus)',
        url: 'Purchase URL',
        urlPlaceholder: 'https://example.com/purchase',
        urlHint: 'Muss eine absolute http(s)-URL sein',
        iframeWarning:
          '⚠️ iframe note: Some websites block embedding via X-Frame-Options or CSP (frame-ancestors). If the page is blank, provide an "Open in new tab" alternative.'
      },
      smtp: {
        title: 'SMTP Settings',
        description: 'E-Mail-Versand fuer Verifizierungscodes konfigurieren',
        testConnection: 'Verbindung testen',
        testing: 'Wird getestet...',
        host: 'SMTP-Host',
        hostPlaceholder: 'smtp.gmail.com',
        port: 'SMTP-Port',
        portPlaceholder: '587',
        username: 'SMTP Username',
        usernamePlaceholder: "your-email{'@'}gmail.com",
        password: 'SMTP Password',
        passwordPlaceholder: '********',
        passwordHint: 'Leer lassen, um bestehendes Passwort beizubehalten',
        passwordConfiguredPlaceholder: '********',
        passwordConfiguredHint: 'Password configured. Leave empty to keep the current value.',
        fromEmail: 'From Email',
        fromEmailPlaceholder: "noreply{'@'}example.com",
        fromName: 'Absendername',
        fromNamePlaceholder: 'Sub2API',
        useTls: 'TLS verwenden',
        useTlsHint: 'Enable TLS encryption for SMTP connection'
      },
      testEmail: {
        title: 'Send Test Email',
        description: 'Test-E-Mail senden, um Ihre SMTP-Konfiguration zu ueberpruefen',
        recipientEmail: 'Recipient Email',
        recipientEmailPlaceholder: "test{'@'}example.com",
        sendTestEmail: 'Send Test Email',
        sending: 'Wird gesendet...',
        enterRecipientHint: 'Bitte geben Sie eine Empfaenger-E-Mail-Adresse ein'

      },
      opsMonitoring: {
        title: 'Ops-Ueberwachung',
        description: 'Enable ops monitoring for troubleshooting and health visibility',
        disabled: 'Ops-Ueberwachung ist deaktiviert',
        enabled: 'Enable Ops Monitoring',
        enabledHint: 'Enable the ops monitoring module (admin only)',
        realtimeEnabled: 'Enable Realtime Monitoring',
        realtimeEnabledHint: 'Enable realtime QPS/metrics push (WebSocket)',
        queryMode: 'Standard-Abfragemodus',
        queryModeHint: 'Standard-Abfragemodus fuer das Ops-Dashboard (auto/raw/preagg)',
        queryModeAuto: 'Auto (empfohlen)',
        queryModeRaw: 'Roh (am genauesten, langsamer)',
        queryModePreagg: 'Voragg (am schnellsten, erfordert Aggregation)',
        metricsInterval: 'Metriken-Erfassungsintervall (Sekunden)',
        metricsIntervalHint: 'Wie oft System-/Anfragemetriken erfasst werden (60-3600 Sekunden)'

      },
      adminApiKey: {
        title: 'Admin API-Schluessel',
        description: 'Globaler API-Schluessel fuer externe Systemintegration mit vollem Admin-Zugang',
        notConfigured: 'Admin API-Schluessel nicht konfiguriert',
        configured: 'Admin API key is active',
        currentKey: 'Aktueller Schluessel',
        regenerate: 'Neu generieren',
        regenerating: 'Wird neu generiert...',
        delete: 'Loeschen',
        deleting: 'Wird geloescht...',
        create: 'Schluessel erstellen',
        creating: 'Wird erstellt...',
        regenerateConfirm: 'Are you sure? The current key will be immediately invalidated.',
        deleteConfirm:
          'Are you sure you want to delete the admin API key? External integrations will stop working.',
        keyGenerated: 'Neuer Admin API-Schluessel generiert',
        keyDeleted: 'Admin API-Schluessel geloescht',
        copyKey: 'Schluessel kopieren',
        keyCopied: 'Schluessel in Zwischenablage kopiert',
        keyWarning: 'Dieser Schluessel wird nur einmal angezeigt. Bitte kopieren Sie ihn jetzt.',
        securityWarning: 'Warning: This key provides full admin access. Keep it secure.',
        usage: 'Usage: Add to request header - x-api-key: <your-admin-api-key>'
      },
      streamTimeout: {
        title: 'Stream Timeout Handling',
        description: 'Kontobehandlungsstrategie konfigurieren, wenn die Upstream-Antwort das Zeitlimit ueberschreitet',
        enabled: 'Enable Stream Timeout Handling',
        enabledHint: 'Problematische Konten automatisch behandeln, wenn Upstream das Zeitlimit ueberschreitet',
        timeoutSeconds: 'Timeout Threshold (seconds)',
        timeoutSecondsHint: 'Stream-Datenintervall, das diese Zeit ueberschreitet, gilt als Zeitlimit (30-300s)',
        action: 'Aktion',
        actionTempUnsched: 'Temporaer nichtplanbar',
        actionError: 'Mark as Error',
        actionNone: 'Keine Aktion',
        actionHint: 'Aktion, die nach dem Zeitlimit auf das Konto angewendet wird',
        tempUnschedMinutes: 'Pause Duration (minutes)',
        tempUnschedMinutesHint: 'Duration of temporary unschedulable state (1-60 minutes)',
        thresholdCount: 'Ausloeser-Schwellenwert (Anzahl)',
        thresholdCountHint: 'Anzahl der Zeitlimits vor Ausloesung der Aktion (1-10)',
        thresholdWindowMinutes: 'Schwellenwert-Fenster (Minuten)',
        thresholdWindowMinutesHint: 'Time window for counting timeouts (1-60 minutes)',
        saved: 'Stream-Zeitlimit-Einstellungen gespeichert',
        saveFailed: 'Failed to save stream timeout settings'
      },
      announcements: {
        title: 'System Announcements',
        description: 'Manage announcements displayed on the console home page',
        titleLabel: 'Ankuendigungstitel',
        titlePlaceholder: 'Ankuendigungsinhalt eingeben',
        dateLabel: 'Datum (optional)',
        datePlaceholder: 'z.B. 2026-01-25',
        add: 'Ankuendigung hinzufuegen',
        empty: 'Keine Ankuendigungen',
        saved: 'Announcements saved',
        saveFailed: 'Failed to save announcements'
      },
      recharge: {
        title: 'Aufladeeinstellungen',
        description: 'Aufladefunktionalitaet und Stufen-Multiplikatoren konfigurieren',
        basicSettings: 'Basic Settings',
        enabled: 'Enable Recharge',
        enabledHint: 'Benutzern erlauben, Kontoguthaben aufzuladen',
        minAmount: 'Mindestbetrag',
        minAmountPlaceholder: '10',
        minAmountHint: 'Mindestaufladebetrag pro Transaktion (¥)',
        maxAmount: 'Hoechstbetrag',
        maxAmountPlaceholder: '10000',
        maxAmountHint: 'Hoechstaufladebetrag pro Transaktion (¥)',
        tiers: 'Tier Multipliers',
        tiersHint: 'Multiplikator-Boni fuer verschiedene Betraege festlegen',
        tier: 'Stufe',
        tierMin: 'Min Amount',
        tierMax: 'Max Amount',
        tierMaxPlaceholder: 'Leer lassen fuer unbegrenzt',
        tierMultiplier: 'Multiplikator',
        addTier: 'Stufe hinzufuegen',
        noTiers: 'Keine Stufen konfiguriert',
        preview: 'Vorschau',
        tierValidation: {
          maxGreaterThanMin: 'Hoechstbetrag muss groesser als Mindestbetrag sein',
          multiplierRange: 'Multiplier must be between 1.0 and 10.0',
          overlap: 'Stufenbereiche duerfen sich nicht ueberschneiden'

        },
        loadFailed: 'Failed to load recharge config',
        saveSuccess: 'Aufladeeinstellungen erfolgreich gespeichert',
        saveFailed: 'Failed to save recharge settings'
      },
      saveSettings: 'Save Settings',
      saving: 'Wird gespeichert...',
      settingsSaved: 'Settings saved successfully',
      smtpConnectionSuccess: 'SMTP-Verbindung erfolgreich',
      testEmailSent: 'Test-E-Mail erfolgreich gesendet',
      failedToLoad: 'Failed to load settings',
      failedToSave: 'Failed to save settings',
      failedToTestSmtp: 'SMTP-Verbindungstest fehlgeschlagen',
      failedToSendTestEmail: 'Failed to send test email'
    },

    // Error Passthrough Rules
    errorPassthrough: {
      title: 'Error Passthrough Rules',
      description: 'Konfigurieren, wie Upstream-Fehler an Clients zurueckgegeben werden',
      createRule: 'Create Rule',
      editRule: 'Edit Rule',
      deleteRule: 'Delete Rule',
      noRules: 'Keine Regeln konfiguriert',
      createFirstRule: 'Create your first error passthrough rule',
      allPlatforms: 'All Platforms',
      passthrough: 'Durchleitung',
      custom: 'Benutzerdefiniert',
      code: 'Code',
      body: 'Body',

      // Columns
      columns: {
        priority: 'Prioritaet',
        name: 'Name',
        conditions: 'Bedingungen',
        platforms: 'Plattformen',
        behavior: 'Verhalten',
        status: 'Status',
        actions: 'Aktionen'

      },

      // Match Mode
      matchMode: {
        any: 'Code ODER Schluesselwort',
        all: 'Code UND Schluesselwort',
        anyHint: 'Statuscode stimmt mit einem Fehlercode ueberein, ODER Nachricht enthaelt ein Schluesselwort',
        allHint: 'Statuscode stimmt mit einem Fehlercode ueberein, UND Nachricht enthaelt ein Schluesselwort'

      },

      // Form
      form: {
        name: 'Regelname',
        namePlaceholder: 'z.B. Kontextlimit-Durchleitung',
        priority: 'Prioritaet',
        priorityHint: 'Niedrigere Werte haben hoehere Prioritaet',
        description: 'Beschreibung',
        descriptionPlaceholder: 'Beschreiben Sie den Zweck dieser Regel...',
        matchConditions: 'Uebereinstimmungsbedingungen',
        errorCodes: 'Error Codes',
        errorCodesPlaceholder: '422, 400, 429',
        errorCodesHint: 'Mehrere Codes durch Kommas trennen',
        keywords: 'Schluesselwoerter',
        keywordsPlaceholder: 'One keyword per line\ncontext limit\nmodel not supported',
        keywordsHint: 'Ein Schluesselwort pro Zeile, Gross-/Kleinschreibung wird nicht beachtet',
        matchMode: 'Uebereinstimmungsmodus',
        platforms: 'Plattformen',
        platformsHint: 'Leer lassen, um auf alle Plattformen anzuwenden',
        responseBehavior: 'Antwortverhalten',
        passthroughCode: 'Upstream-Statuscode durchleiten',
        responseCode: 'Benutzerdefinierter Statuscode',
        passthroughBody: 'Upstream-Fehlermeldung durchleiten',
        customMessage: 'Benutzerdefinierte Fehlermeldung',
        customMessagePlaceholder: 'Error message to return to client...',
        enabled: 'Enable this rule'
      },

      // Messages
      nameRequired: 'Bitte Regelname eingeben',
      conditionsRequired: 'Bitte mindestens einen Fehlercode oder ein Schluesselwort konfigurieren',
      ruleCreated: 'Regel erfolgreich erstellt',
      ruleUpdated: 'Regel erfolgreich aktualisiert',
      ruleDeleted: 'Regel erfolgreich geloescht',
      deleteConfirm: 'Moechten Sie die Regel "{name}" wirklich loeschen?',
      failedToLoad: 'Failed to load rules',
      failedToSave: 'Failed to save rule',
      failedToDelete: 'Failed to delete rule',
      failedToToggle: 'Failed to toggle status'
    }
  },

  // Subscription Progress (Header component)
  subscriptionProgress: {
    title: 'Meine Abonnements',
    viewDetails: 'Abonnementdetails anzeigen',
    activeCount: '{count} active subscription(s)',
    daily: 'Daily',
    weekly: 'Weekly',
    monthly: 'Monthly',
    daysRemaining: 'Noch {days} Tage',
    expired: 'Abgelaufen',
    expiresToday: 'Laeuft heute ab',
    expiresTomorrow: 'Laeuft morgen ab',
    viewAll: 'View all subscriptions',
    noSubscriptions: 'Keine aktiven Abonnements',
    unlimited: 'Unbegrenzt'

  },

  // Version Badge
  version: {
    currentVersion: 'Current Version',
    latestVersion: 'Latest Version',
    upToDate: "Sie verwenden die neueste Version.",
    updateAvailable: 'Eine neue Version ist verfuegbar!',
    releaseNotes: 'Release Notes',
    noReleaseNotes: 'Keine Versionshinweise',
    viewUpdate: 'View Update',
    viewRelease: 'Version anzeigen',
    viewChangelog: 'Aenderungsprotokoll anzeigen',
    refresh: 'Aktualisieren',
    sourceMode: 'Source Build',
    sourceModeHint: 'Quellcode-Build, verwenden Sie git pull zum Aktualisieren',
    updateNow: 'Update Now',
    updating: 'Wird aktualisiert...',
    updateComplete: 'Update Complete',
    updateFailed: 'Update Failed',
    restartRequired: 'Bitte starten Sie den Dienst neu, um das Update anzuwenden',
    restartNow: 'Restart Now',
    restarting: 'Wird neu gestartet...',
    retry: 'Wiederholen'

  },

  // Purchase Subscription Page
  purchase: {
    title: 'Abonnement kaufen',
    description: 'Purchase a subscription via the embedded page',
    openInNewTab: 'Open in new tab',
    notEnabledTitle: 'Feature not enabled',
    notEnabledDesc: 'Der Administrator hat die Kaufseite nicht aktiviert. Bitte kontaktieren Sie den Administrator.',
    notConfiguredTitle: 'Purchase URL not configured',
    notConfiguredDesc:
      'The administrator enabled the entry but has not configured a purchase URL. Please contact admin.'
  },

  // Announcements Page
  announcements: {
    title: 'Ankuendigungen',
    description: 'Systemankuendigungen anzeigen',
    unreadOnly: 'Show unread only',
    markRead: 'Als gelesen markieren',
    markAllRead: 'Alle als gelesen markieren',
    viewAll: 'View all announcements',
    markedAsRead: 'Als gelesen markiert',
    allMarkedAsRead: 'Alle Ankuendigungen als gelesen markiert',
    newCount: '{count} neue Ankuendigung | {count} neue Ankuendigungen',
    readAt: 'Read at',
    read: 'Gelesen',
    unread: 'Ungelesen',
    startsAt: 'Beginnt am',
    endsAt: 'Endet am',
    empty: 'Keine Ankuendigungen',
    emptyUnread: 'Keine ungelesenen Ankuendigungen',
    total: 'Ankuendigungen',
    emptyDescription: 'Es gibt derzeit keine Systemankuendigungen',
    readStatus: 'Sie haben diese Ankuendigung gelesen',
    markReadHint: 'Click "Mark as read" to mark this announcement'
  },

  // User Subscriptions Page
  userSubscriptions: {
    title: 'Meine Abonnements',
    description: 'Ihre Abonnementtarife und Nutzung anzeigen',
    mySubscriptions: 'Meine Abonnements',
    noActiveSubscriptions: 'No Active Subscriptions',
    noActiveSubscriptionsDesc:
      "You don't have any active subscriptions. Contact administrator to get one.",
    purchasePlan: 'Purchase Plan',
    failedToLoad: 'Failed to load subscriptions',
    paymentFailed: 'Zahlung fehlgeschlagen',
    paymentSuccess: 'Zahlung erfolgreich',
    paymentPending: 'Zahlung ausstehend',
    paymentVerifyFailed: 'Zahlungsverifizierung fehlgeschlagen',
    paymentProcessError: 'Zahlungsverarbeitungsfehler',
    status: {
      active: 'Aktiv',
      expired: 'Abgelaufen',
      revoked: 'Widerrufen'

    },
    usage: 'Nutzung',
    expires: 'Laeuft ab',
    noExpiration: 'No expiration',
    unlimited: 'Unbegrenzt',
    unlimitedDesc: 'Keine Nutzungslimits fuer dieses Abonnement',
    daily: 'Daily',
    weekly: 'Weekly',
    monthly: 'Monthly',
    daysRemaining: '{days} days remaining',
    expiresOn: 'Laeuft ab am {date}',
    resetIn: 'Resets in {time}',
    windowNotActive: 'Awaiting first use',
    usageOf: '{used} von {limit}',
    remaining: 'Verbleibend',
    usedOfTotal: 'Used ${used} / Total ${total} ({cycles} cycles)',
    renew: 'Verlaengern',
    renewPrice: 'Verlaengern: ¥{price} / {days} Tage',
    renewError: 'Failed to renew, please try again'
  },

  // Plans (user purchase page)
  plans: {
    title: 'Tarife',
    description: 'Waehlen Sie Ihren Tarif',
    subtitle: 'Waehlen Sie den Tarif, der Ihren Beduerfnissen entspricht',
    loadError: 'Failed to load plans',
    noPlans: 'No Plans Available',
    noPlansDesc: 'Es sind keine kaufbaren Tarife konfiguriert. Bitte schauen Sie spaeter noch einmal vorbei.',
    recommended: 'Empfohlen',
    validityPeriod: '{days} Tage Gueltigkeit',
    validFor: 'Gueltig fuer {days} Tage',
    dailyLimit: 'Daily limit ${amount}',
    weeklyLimit: 'Weekly limit ${amount}',
    monthlyLimit: 'Monthly limit ${amount}',
    unlimited: 'Unbegrenzt',
    processing: 'Wird verarbeitet...',
    purchase: 'Purchase Now',
    purchaseError: 'Failed to create order. Please try again.',
    buyOnTaobao: 'Auf Taobao kaufen',
    pricePerDollar: 'Nur ¥{price}/Dollar',
    estimatedUsage: '~{count} Aufrufe (gemischte Nutzung)',
    estimatedTokens: '~{tokens} Tokens',
    paygo: {
      title: 'PayGo Nutzungsbasiert',
      description: 'Flexibles Aufladen, nutzen nach Bedarf',
      features: {
        anyAmount: 'Beliebigen Betrag aufladen',
        payAsYouGo: 'Zahlen Sie nur fuer das, was Sie nutzen',
        neverExpires: 'Guthaben laeuft nie ab'

      }
    },
    cryptoBanner: {
      title: 'Kryptowährungszahlung verfügbar',
      descriptionPrefix: 'Kontaktieren Sie uns nach der Überweisung über',
      descriptionSuffix: ', um Ihre Bestellung zu bestätigen',
      copySuccess: 'Adresse kopiert'
    },
    enterprise: {
      badge: 'Premium',
      name: 'Enterprise',
      description: 'Dedizierte Leitung, ultimatives Erlebnis, massgeschneidert fuer Unternehmenskunden',
      contactUs: 'Kontaktieren Sie uns',
      customized: 'Massgeschneiderte Loesungen',
      contact: 'Contact Sales',
      contactMessage: 'Please contact us for a custom plan: WeChat AI码驿站',
      features: {
        dedicated: 'Dedizierte private Leitung',
        unlimited: 'Unlimited quota',
        priority: 'Prioritaets-Support',
        sla: '99,9% SLA-Garantie'

      },
      dialog: {
        title: 'Kontaktieren Sie uns',
        subtitle: 'QR-Code scannen, um WeChat fuer massgeschneiderte Loesungen hinzuzufuegen',
        tip: 'Geschaeftszeiten: Mo-So 9:00-22:00'

      }
    }
  },

  // Recharge
  recharge: {
    title: 'Recharge Account',
    rechargeNow: 'Jetzt aufladen',
    rechargeAmount: 'Aufladebetrag (¥)',
    enterAmount: 'Aufladebetrag eingeben',
    quickAmounts: 'Quick Amounts',
    currentBalance: 'Current Balance',
    bonus: 'Aufladebonus',
    actualCredit: 'Actual Credit',
    multiplier: 'Multiplikator',
    multiplierInfo: '{multiplier}× Multiplikator',
    tierInfo: 'Recharge Tiers',
    bonusTip: 'Mehr aufladen, mehr Bonus erhalten!',
    minAmount: 'Min Amount',
    maxAmount: 'Max Amount',
    confirmRecharge: 'Confirm Recharge',
    processing: 'Wird verarbeitet...',
    rechargeFailed: 'Aufladen fehlgeschlagen',
    invalidAmount: 'Betrag muss zwischen ¥{min} und ¥{max} liegen',
    rechargeSuccess: 'Aufladen erfolgreich',

    // Promotional copy
    promoTitle: '💰 Recharge for bonus! More recharge, more bonus',
    promoSubtitle: 'Balance never expires, pay as you use, great value',
    quickTip: 'Recommended: ¥200+ for better value',
    moreGetMore: 'More recharge, more bonus',
    usageRuleTitle: '💡 Usage Rules',
    usageRuleDesc: 'Balance is consumed at ¥1 = $1. Higher multiplier = lower unit price',
    benefitTitle: '✨ Balance never expires',
    benefitDesc: 'Aufgeladenes Guthaben ist dauerhaft, Abrechnung nach tatsaechlicher Nutzung, keine Verschwendung',

    // Order related
    myOrders: 'Aufladeverlauf',
    orderNo: 'Order No.',
    amount: 'Payment Amount',
    creditAmount: 'Credit Amount',
    status: 'Status',
    createdAt: 'Erstellt am',
    paidAt: 'Paid At',
    continuePay: 'Continue Payment',
    paying: 'Wird verarbeitet...',
    expired: 'Abgelaufen',
    statusLabels: {
      pending: 'Ausstehend',
      paid: 'Bezahlt',
      expired: 'Abgelaufen',
      refunded: 'Refunded'
    },
    noOrders: 'Keine Aufladeeintraege',
    noOrdersDesc: "Sie haben noch nicht aufgeladen",
    failedToLoad: 'Failed to load recharge records',
    repayFailed: 'Failed to get payment link'
  },

  // User Orders
  userOrders: {
    title: 'Meine Bestellungen',
    description: 'Ihren Bestellverlauf anzeigen',
    myOrders: 'Meine Bestellungen',
    noOrders: 'Keine Bestellungen',
    noOrdersDesc: "Sie haben noch keine Tarife gekauft",
    purchasePlan: 'Purchase a Plan',
    orderNo: 'Order No.',
    planName: 'Tarif',
    amount: 'Betrag',
    status: 'Status',
    createdAt: 'Erstellt am',
    failedToLoad: 'Failed to load orders',
    statusLabels: {
      pending: 'Ausstehend',
      paid: 'Bezahlt',
      expired: 'Abgelaufen',
      refunded: 'Refunded'
    },
    actions: 'Aktionen',
    continuePay: 'Pay Now',
    paying: 'Wird verarbeitet...',
    expired: 'Abgelaufen',
    repayFailed: 'Failed to get payment link',
    planOrders: 'Tarifbestellungen',
    rechargeOrders: 'Aufladebestellungen'

  },

  // Onboarding Tour
  onboarding: {
    restartTour: 'Restart Onboarding Tour',
    dontShowAgain: "Nicht mehr anzeigen",
    dontShowAgainTitle: 'Einfuehrungsanleitung dauerhaft schliessen',
    confirmDontShow: "Are you sure you don't want to see the onboarding guide again?\n\nYou can restart it anytime from the user menu in the top right corner.",
    confirmExit: 'Are you sure you want to exit the onboarding guide? You can restart it anytime from the top right menu.',
    interactiveHint: 'Druecken Sie Enter oder klicken Sie, um fortzufahren',
    navigation: {
      flipPage: 'Flip Page',
      exit: 'Beenden'

    },
    // Admin tour steps
    admin: {
      welcome: {
        title: '👋 Welcome to Sub2API',
        description: '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Sub2API is a powerful AI service gateway platform that helps you easily manage and distribute AI services.</p><p style="margin-bottom: 12px;"><b>🎯 Core Features:</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>📦 <b>Group Management</b> - Create service tiers (VIP, Free Trial, etc.)</li><li>🔗 <b>Account Pool</b> - Connect multiple upstream AI service accounts</li><li>🔑 <b>Key Distribution</b> - Generate independent API Keys for users</li><li>💰 <b>Billing Control</b> - Flexible rate and quota management</li></ul><p style="color: #10b981; font-weight: 600;">Let\'s complete the initial setup in 3 minutes →</p></div>',
        nextBtn: 'Start Setup 🚀',
        prevBtn: 'Ueberspringen'

      },
      groupManage: {
        title: '📦 Step 1: Group Management',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>What is a Group?</b></p><p style="margin-bottom: 12px;">Groups are the core concept of Sub2API, like a "service package":</p><ul style="margin-left: 20px; margin-bottom: 12px; font-size: 13px;"><li>🎯 Each group can contain multiple upstream accounts</li><li>💰 Each group has independent billing multiplier</li><li>👥 Can be set as public or exclusive</li></ul><p style="margin-top: 12px; padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Example:</b> You can create "VIP Premium" (high rate) and "Free Trial" (low rate) groups</p><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 Click "Group Management" on the left sidebar</p></div>'
      },
      createGroup: {
        title: '➕ Create New Group',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Let\'s create your first group.</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📝 Tip:</b> Recommend creating a test group first to familiarize yourself with the process</p><p style="color: #10b981; font-weight: 600;">👉 Click the "Create Group" button</p></div>'
      },
      groupName: {
        title: '✏️ 1. Group Name',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Give your group an easy-to-identify name.</p><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>💡 Naming Suggestions:</b><ul style="margin: 8px 0 0 16px;"><li>"Test Group" - For testing</li><li>"VIP Premium" - High-quality service</li><li>"Free Trial" - Trial version</li></ul></div><p style="font-size: 13px; color: #6b7280;">Click "Next" when done</p></div>',
        nextBtn: 'Weiter'

      },
      groupPlatform: {
        title: '🤖 2. Select Platform',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Choose the AI platform this group supports.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📌 Platform Guide:</b><ul style="margin: 8px 0 0 16px;"><li><b>Anthropic</b> - Claude models</li><li><b>OpenAI</b> - GPT models</li><li><b>Google</b> - Gemini models</li></ul></div><p style="font-size: 13px; color: #6b7280;">One group can only have one platform</p></div>',
        nextBtn: 'Weiter'

      },
      groupMultiplier: {
        title: '💰 3. Rate Multiplier',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Set the billing multiplier to control user charges.</p><div style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚙️ Billing Rules:</b><ul style="margin: 8px 0 0 16px;"><li><b>1.0</b> - Original price (cost price)</li><li><b>1.5</b> - User consumes $1, charged $1.5</li><li><b>2.0</b> - User consumes $1, charged $2</li><li><b>0.8</b> - Subsidy mode (loss-making)</li></ul></div><p style="font-size: 13px; color: #6b7280;">Recommend setting test group to 1.0</p></div>',
        nextBtn: 'Weiter'

      },
      groupExclusive: {
        title: '🔒 4. Exclusive Group (Optional)',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Control group visibility and access permissions.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🔐 Permission Guide:</b><ul style="margin: 8px 0 0 16px;"><li><b>Off</b> - Public group, visible to all users</li><li><b>On</b> - Exclusive group, only for specified users</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Use Cases:</b> VIP exclusive, internal testing, special customers</p></div>',
        nextBtn: 'Weiter'

      },
      groupSubmit: {
        title: '✅ Save Group',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Confirm the information and click create to save the group.</p><p style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ Note:</b> Platform type cannot be changed after creation, but other settings can be edited anytime</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>📌 Next Step:</b> After creation, we\'ll add upstream accounts to this group</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create" button</p></div>'
      },
      accountManage: {
        title: '🔗 Step 2: Add Account',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>Great! Group created successfully 🎉</b></p><p style="margin-bottom: 12px;">Now add upstream AI service accounts to enable actual service delivery.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🔑 Account Purpose:</b><ul style="margin: 8px 0 0 16px;"><li>Connect to upstream AI services (Claude, GPT, etc.)</li><li>One group can contain multiple accounts (load balancing)</li><li>Supports OAuth and Session Key methods</li></ul></div><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 Click "Account Management" on the left sidebar</p></div>'
      },
      createAccount: {
        title: '➕ Add New Account',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Click the button to start adding your first upstream account.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> Recommend using OAuth method - more secure and no manual key extraction needed</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Add Account" button</p></div>'
      },
      accountName: {
        title: '✏️ 1. Account Name',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Set an easy-to-identify name for the account.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Naming Suggestions:</b> "Claude Main", "GPT Backup 1", "Test Account", etc.</p></div>',
        nextBtn: 'Weiter'

      },
      accountPlatform: {
        title: '🤖 2. Select Platform',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Choose the service provider platform for this account.</p><p style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px;"><b>⚠️ Important:</b> Platform must match the group you just created</p></div>',
        nextBtn: 'Weiter'

      },
      accountType: {
        title: '🔐 3. Authorization Method',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Choose the account authorization method.</p><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>✅ Recommended: OAuth Method</b><ul style="margin: 8px 0 0 16px;"><li>No manual key extraction needed</li><li>More secure with auto-refresh support</li><li>Works with Claude Code, ChatGPT OAuth</li></ul></div><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 Session Key Method</b><ul style="margin: 8px 0 0 16px;"><li>Requires manual extraction from browser</li><li>May need periodic updates</li><li>For platforms without OAuth support</li></ul></div></div>',
        nextBtn: 'Weiter'

      },
      accountPriority: {
        title: '⚖️ 4. Priority (Optional)',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Set the account call priority.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📊 Priority Rules:</b><ul style="margin: 8px 0 0 16px;"><li>Lower number = higher priority</li><li>System uses low-value accounts first</li><li>Same priority = random selection</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Use Case:</b> Set main account to lower value, backup accounts to higher value</p></div>',
        nextBtn: 'Weiter'

      },
      accountGroups: {
        title: '🎯 5. Assign Groups',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>Key Step!</b> Assign the account to the group you just created.</p><div style="padding: 8px 12px; background: #fee2e2; border-left: 3px solid #ef4444; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ Important Reminder:</b><ul style="margin: 8px 0 0 16px;"><li>Must select at least one group</li><li>Unassigned accounts cannot be used</li><li>One account can be assigned to multiple groups</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> Select the test group you just created</p></div>',
        nextBtn: 'Weiter'

      },
      accountSubmit: {
        title: '✅ Save Account',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Confirm the information and click save.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📌 OAuth Flow:</b><ul style="margin: 8px 0 0 16px;"><li>Will redirect to service provider page after clicking save</li><li>Complete login and authorization on provider page</li><li>Auto-return after successful authorization</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>📌 Next Step:</b> After adding account, we\'ll create an API key</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Save" button</p></div>'
      },
      keyManage: {
        title: '🔑 Step 3: Generate Key',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>Congratulations! Account setup complete 🎉</b></p><p style="margin-bottom: 12px;">Final step: generate an API Key to test if the service works properly.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🔑 API Key Purpose:</b><ul style="margin: 8px 0 0 16px;"><li>Credential for calling AI services</li><li>Each key is bound to one group</li><li>Can set quota and expiration</li><li>Supports independent usage statistics</li></ul></div><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 Click "API Keys" on the left sidebar</p></div>'
      },
      createKey: {
        title: '➕ Create Key',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Click the button to create your first API Key.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> Copy and save immediately after creation - key is only shown once</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create Key" button</p></div>'
      },
      keyName: {
        title: '✏️ 1. Key Name',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Set an easy-to-manage name for the key.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Naming Suggestions:</b> "Test Key", "Production", "Mobile", etc.</p></div>',
        nextBtn: 'Weiter'

      },
      keyGroup: {
        title: '🎯 2. Select Group',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Select the group you just configured.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📌 Group Determines:</b><ul style="margin: 8px 0 0 16px;"><li>Which accounts this key can use</li><li>What billing multiplier applies</li><li>Whether it\'s an exclusive key</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> Select the test group you just created</p></div>',
        nextBtn: 'Weiter'

      },
      keySubmit: {
        title: '🎉 Generate and Copy',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">System will generate a complete API Key after clicking create.</p><div style="padding: 8px 12px; background: #fee2e2; border-left: 3px solid #ef4444; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ Important Reminder:</b><ul style="margin: 8px 0 0 16px;"><li>Key is only shown once, copy immediately</li><li>Need to regenerate if lost</li><li>Keep it safe, don\'t share with others</li></ul></div><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🚀 Next Steps:</b><ul style="margin: 8px 0 0 16px;"><li>Copy the generated sk-xxx key</li><li>Use in any OpenAI-compatible client</li><li>Start experiencing AI services!</li></ul></div><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create" button</p></div>'
      }
    },
    // User tour steps
    user: {
      welcome: {
        title: '👋 Welcome to Sub2API',
        description: '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Hello! Welcome to the Sub2API AI service platform.</p><p style="margin-bottom: 12px;"><b>🎯 Quick Start:</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>🔑 Create API Key</li><li>📋 Copy key to your application</li><li>🚀 Start using AI services</li></ul><p style="color: #10b981; font-weight: 600;">Just 1 minute, let\'s get started →</p></div>',
        nextBtn: 'Start 🚀',
        prevBtn: 'Ueberspringen'

      },
      keyManage: {
        title: '🔑 API Key Management',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Manage all your API access keys here.</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 What is an API Key?</b><br/>An API key is your credential for accessing AI services, like a key that allows your application to call AI capabilities.</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click to enter key page</p></div>'
      },
      createKey: {
        title: '➕ Create New Key',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Click the button to create your first API key.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> Key is only shown once after creation, make sure to copy and save</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create Key"</p></div>'
      },
      keyName: {
        title: '✏️ Key Name',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Give your key an easy-to-identify name.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Examples:</b> "My First Key", "For Testing", etc.</p></div>',
        nextBtn: 'Weiter'

      },
      keyGroup: {
        title: '🎯 Select Group',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Select the service group assigned by the administrator.</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 Group Info:</b><br/>Different groups may have different service quality and billing rates, choose according to your needs.</p></div>',
        nextBtn: 'Weiter'

      },
      keySubmit: {
        title: '🎉 Complete Creation',
        description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Click to confirm and create your API key.</p><div style="padding: 8px 12px; background: #fee2e2; border-left: 3px solid #ef4444; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ Important:</b><ul style="margin: 8px 0 0 16px;"><li>Copy the key (sk-xxx) immediately after creation</li><li>Key is only shown once, need to regenerate if lost</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>🚀 How to Use:</b><br/>Configure the key in any OpenAI-compatible client (like ChatBox, OpenCat, etc.) and start using!</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create" button</p></div>'
      }
    }
  },
  keyQuery: {
    title: 'API Key Abfrage',
    placeholder: 'Geben Sie Ihren API Key ein',
    submit: 'Abfragen',
    querying: 'Abfrage...',
    invalidKey: 'Ungultiger API Key',
    keyInfo: 'Schlusselinformationen',
    keyName: 'Name',
    status: 'Status',
    quota: 'Kontingent',
    used: 'Verwendet',
    remaining: 'Verbleibend',
    unlimited: 'Unbegrenzt',
    expiresAt: 'Ablaufdatum',
    neverExpires: 'Lauft nie ab',
    daysRemaining: 'Lauft in {days} Tagen ab',
    subscription: 'Abonnement',
    noSubscription: 'Kein aktives Abonnement',
    dailyUsage: 'Tagesverbrauch',
    weeklyUsage: 'Wochenverbrauch',
    monthlyUsage: 'Monatsverbrauch',
    usageSummary: 'Nutzungsubersicht (7 Tage)',
    totalCost: 'Gesamtkosten',
    requestCount: 'Anfragen',
    topModels: 'Top-Modelle',
    statusActive: 'Aktiv',
    statusInactive: 'Inaktiv',
    statusExpired: 'Abgelaufen',
    statusExhausted: 'Kontingent erschopft',
    noLimit: 'Kein Limit',
    requests: 'Anfragen',
    tabs: {
      overview: 'Ubersicht',
      usageDetails: 'Nutzungsdetails'
    },
    timeRange: {
      today: 'Heute',
      thisMonth: 'Dieser Monat',
      all: 'Alle',
      custom: 'Benutzerdefiniert'
    },
    stats: {
      totalTokens: 'Gesamte Tokens',
      inputTokens: 'Eingabe-Tokens',
      outputTokens: 'Ausgabe-Tokens',
      cacheTokens: 'Cache-Tokens',
      avgDuration: 'Durchschn. Dauer'
    },
    models: {
      title: 'Modellnutzung',
      requests: 'Anfragen',
      tokens: 'Tokens',
      cost: 'Kosten'
    },
    usage: {
      time: 'Zeit',
      model: 'Modell',
      inputTokens: 'Eingabe',
      outputTokens: 'Ausgabe',
      cacheTokens: 'Cache',
      cost: 'Kosten',
      duration: 'Dauer',
      noData: 'Keine Daten',
      prev: 'Zuruck',
      next: 'Weiter'
    },
    trend: {
      title: 'Nutzungstrend'
    },
    language: 'Sprache'
  },
  reseller: {
    dashboard: {
      title: 'Händler-Dashboard',
      myBalance: 'Mein Guthaben',
      domains: 'Eigene Domains',
      verifiedDomains: 'Verifizierte Domains',
      groups: 'Pakete',
      keyCount: 'Schlüssel gesamt',
      activeKeys: 'Aktive Schlüssel',
      totalQuotaUsed: 'Verbrauchtes Kontingent',
      quickActions: 'Schnellaktionen',
      manageDomains: 'Domains verwalten',
      manageDomainsDesc: 'Eigene Domains und Branding konfigurieren',
      apiKeys: 'API-Schlüssel',
      apiKeysDesc: 'API-Schlüssel verwalten'
    },
    keys: {
      title: 'API-Schlüssel-Verwaltung',
      create: 'Schlüssel erstellen',
      edit: 'Schlüssel bearbeiten',
      delete: 'Schlüssel löschen',
      name: 'Name',
      notes: 'Notizen',
      key: 'Schlüssel',
      group: 'Paket',
      status: 'Status',
      quota: 'Kontingent',
      quotaUsed: 'Verbrauchtes Kontingent',
      expiresAt: 'Läuft ab am',
      ipWhitelist: 'IP-Whitelist',
      ipBlacklist: 'IP-Blacklist',
      customKey: 'Benutzerdefinierter Schlüssel',
      expiresInDays: 'Gültigkeitstage',
      resetQuota: 'Kontingent zurücksetzen',
      copyKey: 'Schlüssel kopieren',
      noKeys: 'Noch keine API-Schlüssel',
      confirmDelete: 'Möchten Sie diesen Schlüssel wirklich löschen?',
      actions: 'Aktionen',
      active: 'Aktiv',
      disabled: 'Deaktiviert',
      unlimited: 'Unbegrenzt',
      noExpiry: 'Kein Ablauf',
      createSuccess: 'Schlüssel erfolgreich erstellt',
      updateSuccess: 'Schlüssel erfolgreich aktualisiert',
      deleteSuccess: 'Schlüssel gelöscht',
      resetSuccess: 'Kontingent erfolgreich zurückgesetzt',
      keyCopied: 'Schlüssel in die Zwischenablage kopiert'
    },
    domains: {
      title: 'Eigene Domains',
      description: 'Eigene Domains konfigurieren und Ihre Marke aufbauen',
      add: 'Domain hinzufügen',
      addTitle: 'Domain hinzufügen',
      editTitle: 'Domain bearbeiten',
      domainName: 'Domainname',
      siteName: 'Seitenname',
      siteLogo: 'Seiten-Logo-URL',
      brandColor: 'Markenfarbe',
      verified: 'Verifiziert',
      unverified: 'Nicht verifiziert',
      verify: 'Verifizieren',
      verifyNow: 'Jetzt verifizieren',
      verifying: 'Wird verifiziert...',
      verifyInstructions: 'Bitte fügen Sie den folgenden DNS-TXT-Eintrag hinzu, um den Domainbesitz zu bestätigen:',
      verifyStep1_dns: 'Schritt 1: A-Eintrag hinzufügen (Domain auf Server zeigen)',
      verifyStep1_hint: `HOST {'@'} steht für die Root-Domain. Wenn Ihr DNS-Anbieter das Domain-Suffix automatisch anhängt, geben Sie einfach {'@'} ein.`,
      verifyStep2_txt: 'Schritt 2: TXT-Eintrag hinzufügen (Domain-Eigentum verifizieren)',
      verifyStep2_hint: 'HOST nur _domain-verify eingeben, ohne das Domain-Suffix.',
      verifyStep3: 'Schritt 3: Warten Sie auf die DNS-Propagierung (meist wenige Minuten) und klicken Sie dann auf "Jetzt verifizieren"',
      verifySuccess: 'Domain-Verifizierung erfolgreich',
      verifyFailed: 'Domain-Verifizierung fehlgeschlagen. Bitte überprüfen Sie die DNS-Konfiguration',
      createSuccess: 'Domain hinzugefügt',
      deleteSuccess: 'Domain gelöscht',
      deleteConfirm: 'Möchten Sie die Domain {domain} wirklich löschen?',
      empty: 'Keine eigenen Domains vorhanden',
      subtitle: 'Subtitle',
      homeContent: 'Home Content',
      homeContentPlaceholder: 'Supports HTML',
      docUrl: 'Documentation URL',
      customCSS: 'Custom CSS',
      customCSSPlaceholder: 'Enter custom CSS styles'
    },
    sites: {
      title: 'Website-Verwaltung',
      description: 'Verwalten Sie Ihre Websites, Domains, Marken und Startseiten',
      add: 'Website hinzufügen',
      addTitle: 'Website hinzufügen',
      backToList: 'Zurück zur Liste',
      editSite: 'Website-Einstellungen',
      tabs: { basic: 'Grundlagen', homepage: 'Startseite', docs: 'Dokumentation', purchase: 'Kauf', seo: 'SEO' },
      domainName: 'Domain',
      siteName: 'Website-Name',
      siteNamePlaceholder: 'Website-Name eingeben',
      subtitle: 'Untertitel',
      subtitlePlaceholder: 'Untertitel eingeben',
      siteLogo: 'Logo-URL',
      siteLogoPlaceholder: 'Logo-Bild-URL eingeben',
      homeTemplate: 'Startseiten-Vorlage',
      homeTemplateDefault: 'Systemstandard',
      homeTemplateMinimal: 'Minimal',
      homeTemplateCustom: 'Benutzerdefiniertes HTML',
      homeTemplateExternal: 'Externe URL',
      homeTemplateDefaultHint: 'Verwendet die vom Systemadministrator konfigurierte Startseite',
      homeTemplateMinimalHint: 'Einfache Startseite mit nur Website-Name und Login',
      homeContent: 'Startseiten-Inhalt',
      homeContentHtmlPlaceholder: 'HTML-Inhalt eingeben',
      homeContentUrlPlaceholder: 'Externe Seiten-URL eingeben (wird als iframe eingebettet)',
      preview: 'Website-Vorschau',
      defaultLocale: 'Standardsprache',
      defaultLocaleBrowser: 'Browser folgen',
      defaultLocaleHint: 'Besucher dieser Website sehen standardmäßig diese Sprache',
      docUrl: 'Dokumentations-URL',
      docUrlPlaceholder: 'Dokumentationsseiten-URL eingeben',
      purchaseEnabled: 'Kaufseite aktivieren',
      purchaseEnabledHint: 'Wenn aktiviert, können Benutzer Abonnements kaufen',
      purchaseUrl: 'Kaufseiten-URL',
      purchaseUrlPlaceholder: 'Benutzerdefinierte Kaufseiten-URL eingeben',
      purchaseUrlHint: 'Leer lassen für Standard-Kaufseite',
      seoSection: 'SEO',
      seoTitle: 'SEO-Titel',
      seoTitlePlaceholder: 'Suchmaschinen-Titel',
      seoDescription: 'SEO-Beschreibung',
      seoDescriptionPlaceholder: 'Suchmaschinen-Beschreibung',
      seoKeywords: 'SEO-Schlüsselwörter',
      seoKeywordsPlaceholder: 'Mit Komma trennen',
      empty: 'Noch keine Websites',
      verificationStatus: 'Verifizierungsstatus'
    },
    groups: {
      title: 'Paketverwaltung',
      description: 'Pakete erstellen und verwalten',
      create: 'Paket erstellen',
      createTitle: 'Paket erstellen',
      editTitle: 'Paket bearbeiten',
      template: 'Paketvorlage',
      selectTemplate: 'Vorlage auswählen...',
      price: 'Preis',
      validityDays: 'Gültigkeitstage',
      dailyLimit: 'Tageslimit',
      weeklyLimit: 'Wochenlimit',
      monthlyLimit: 'Monatslimit',
      noLimit: 'Kein Limit',
      unlimited: 'Unbegrenzt',
      purchasable: 'Kaufbar',
      createSuccess: 'Paket erfolgreich erstellt',
      deleteSuccess: 'Paket gelöscht',
      deleteConfirm: 'Paket „{name}" wirklich löschen?',
      empty: 'Keine Pakete vorhanden'
    },
    settings: {
      title: 'Einstellungen',
      description: 'Händler-Website konfigurieren',
      brand: 'Markeneinstellungen',
      siteName: 'Seitenname',
      siteLogo: 'Seiten-Logo URL',
      siteSubtitle: 'Seiten-Untertitel',
      brandColor: 'Markenfarbe',
      customCSS: 'Benutzerdefiniertes CSS',
      customCSSPlaceholder: 'CSS-Code eingeben...',
      homeContent: 'Startseiten-Inhalt',
      homeContentPlaceholder: 'Benutzerdefinierten Inhalt eingeben (HTML unterstützt)...',
      registration: 'Registrierungseinstellungen',
      registrationEnabled: 'Offene Registrierung',
      emailVerifyEnabled: 'E-Mail-Verifizierung',
      invitationCodeEnabled: 'Einladungscode erforderlich',
      defaultBalance: 'Standardguthaben',
      defaultConcurrency: 'Standard-Parallelität',
      defaultLocale: 'Standardsprache',
      defaultLocaleBrowser: 'Browser folgen',
      defaultLocaleHint: 'Benutzer, die über Ihre Domain zugreifen, sehen standardmäßig diese Sprache',
      other: 'Sonstige Einstellungen',
      docUrl: 'Dokumentations-URL',
      contactInfo: 'Kontaktinformationen',
      cryptoAddresses: 'Krypto-Adressen',
      cryptoPlaceholder: 'Eine pro Zeile, Format: COIN:Adresse',
      telegram: {
        title: 'Telegram Bot',
        botToken: 'Bot-Token',
        botTokenPlaceholder: `Token von {'@'}BotFather eingeben`,
        botTokenHint: `Erhalten Sie es von Telegram {'@'}BotFather`,
        bindStatus: 'Bindungsstatus',
        bound: 'Gebunden',
        unbound: 'Nicht gebunden',
        generateBindCode: 'Bindungscode generieren',
        bindInstructions: 'Senden Sie diesen Befehl an Ihren Bot zum Binden:',
        unbind: 'Bindung aufheben',
        saveTokenFirst: 'Bitte Bot Token eingeben und Einstellungen zuerst speichern',
        waitingForBind: 'Warte auf Bindung, bitte senden Sie den Befehl an Ihren Bot...',
        bindSuccess: 'Telegram erfolgreich gebunden!'
      }
    }
  }
}
