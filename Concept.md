🛑 Le Concept en bref

Une plateforme cross-platform (Web + Mobile) synchronisée en temps réel, basée sur la géolocalisation, qui connecte le surplus alimentaire des restaurateurs avec les besoins des animaux errants via une communauté active et récompensée par la gamification.
👥 Les Types d'Utilisateurs

Pour que l'application fonctionne, elle repose sur deux rôles principaux (soumis à une inscription obligatoire) :

    Les Givers (Donateurs) : Principalement des restaurateurs. Ils publient des annonces pour donner leurs invendus ou de la nourriture périmée/à jeter mais encore consommable pour les animaux.

    Les Feeders (Nourrisseurs) : Les utilisateurs qui localisent les animaux errants, récupèrent la nourriture chez les Givers et vont nourrir les animaux sur les points stratégiques.

🗺️ Les Fonctionnalités Clés
1. La Carte Interactive (Cœur de l'application)

    Pings Animaux : Signalement en temps réel des chiens/chats errants aperçus.

    Pings Restos : Localisation des points de collecte de nourriture disponibles chez les Givers.

    Historique et Activité du Pin : Au clic sur un marqueur, accès à un fil d'actualité (L'animal est-il toujours là ? A-t-il été nourri récemment ?).

    Preuve Sociale : Possibilité d'ajouter des photos/vidéos pour prouver l'action de nourrissage et actualiser le statut du ping.

2. Gamification & Engagement

    Système d'XP : Chaque action (signaler un animal, donner de la nourriture, confirmer la présence, uploader une photo) rapporte des points d'expérience.

    Badges de Certification : Des badges évolutifs pour valider le statut et la fiabilité des utilisateurs au sein de la communauté.

💰 Modèle Économique (Monétisation)

L'application est gratuite, mais intègre trois leviers financiers pour s'autofinancer (serveurs, API de cartographie, etc.) :

    Publicités : Affichées pour les utilisateurs de la version gratuite.

    Abonnement Premium (Optionnel) : Supprime les publicités et offre un badge exclusif "Premium ++" sur le profil.

    Dons : Un système de tracking/soutien financier direct pour les utilisateurs qui souhaitent aider sans forcément être sur le terrain.

🛠️ Stack Technique Envisagée

    Backend (Golang) : Parfait pour gérer le temps réel, les connexions simultanées, la synchronisation Web/Mobile et le calcul de positions géographiques (avec une base de données comme PostgreSQL + PostGIS).

    Frontend (React / React Native) : Idéal pour réutiliser la logique de code entre le site Web (React) et l'application mobile (React Native), assurant une synchronisation parfaite de la carte et des données.