CREATE TABLE "public"."deployment_events" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"(),
    "user_id" "uuid",
    "deployment_id" "uuid",
    "message" "text",
    "event_type" character varying
);
