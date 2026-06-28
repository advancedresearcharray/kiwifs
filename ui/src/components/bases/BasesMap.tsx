// Map layout for the Bases component.
// Uses react-map-gl with MapLibre GL and OpenStreetMap tiles (no API key).

import { useCallback, useMemo, useState } from "react";
import Map, { Marker, Popup, NavigationControl } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import type { ViewRow } from "./BasesTable";

type Props = {
  data: ViewRow[];
  onNavigate: (path: string) => void;
};

type MarkerData = {
  path: string;
  title: string;
  lat: number;
  lng: number;
  props: Record<string, unknown>;
};

function extractMarkers(data: ViewRow[]): MarkerData[] {
  const markers: MarkerData[] = [];
  for (const row of data) {
    let lat: number | null = null;
    let lng: number | null = null;

    // Try explicit latitude/longitude fields
    if (row.latitude != null && row.longitude != null) {
      lat = Number(row.latitude);
      lng = Number(row.longitude);
    } else if (row.lat != null && row.lng != null) {
      lat = Number(row.lat);
      lng = Number(row.lng);
    } else if (typeof row.location === "string") {
      // Try "lat,lng" format
      const parts = row.location.split(",").map((s) => parseFloat(s.trim()));
      if (parts.length === 2 && !isNaN(parts[0]!) && !isNaN(parts[1]!)) {
        lat = parts[0]!;
        lng = parts[1]!;
      }
    }

    if (lat != null && lng != null && !isNaN(lat) && !isNaN(lng)) {
      markers.push({ path: row.path, title: row.title, lat, lng, props: row });
    }
  }
  return markers;
}

function fitBounds(markers: MarkerData[]): {
  latitude: number;
  longitude: number;
  zoom: number;
} {
  if (markers.length === 0) {
    return { latitude: 20, longitude: 0, zoom: 2 };
  }
  if (markers.length === 1) {
    return { latitude: markers[0]!.lat, longitude: markers[0]!.lng, zoom: 12 };
  }
  const lats = markers.map((m) => m.lat);
  const lngs = markers.map((m) => m.lng);
  const minLat = Math.min(...lats);
  const maxLat = Math.max(...lats);
  const minLng = Math.min(...lngs);
  const maxLng = Math.max(...lngs);
  const centerLat = (minLat + maxLat) / 2;
  const centerLng = (minLng + maxLng) / 2;
  const latDiff = maxLat - minLat;
  const lngDiff = maxLng - minLng;
  const maxDiff = Math.max(latDiff, lngDiff, 0.01);
  // Rough zoom estimation
  const zoom = Math.max(1, Math.min(14, Math.floor(7 - Math.log2(maxDiff))));
  return { latitude: centerLat, longitude: centerLng, zoom };
}

export function BasesMap({ data, onNavigate }: Props) {
  const markers = useMemo(() => extractMarkers(data), [data]);
  const initialView = useMemo(() => fitBounds(markers), [markers]);
  const [selected, setSelected] = useState<MarkerData | null>(null);

  const handleMarkerClick = useCallback((m: MarkerData) => {
    setSelected(m);
  }, []);

  if (markers.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
        <div className="text-center">
          <p>No geographic data found.</p>
          <p className="text-xs mt-1">
            Add <code className="bg-muted px-1 rounded">latitude</code> and{" "}
            <code className="bg-muted px-1 rounded">longitude</code> frontmatter
            properties to your pages.
          </p>
        </div>
      </div>
    );
  }

  return (
    <Map
      initialViewState={initialView}
      style={{ width: "100%", height: "100%" }}
      mapStyle={{
        version: 8,
        sources: {
          osm: {
            type: "raster",
            tiles: ["https://tile.openstreetmap.org/{z}/{x}/{y}.png"],
            tileSize: 256,
            attribution: "&copy; OpenStreetMap contributors",
          },
        },
        layers: [
          {
            id: "osm",
            type: "raster",
            source: "osm",
          },
        ],
      }}
    >
      <NavigationControl position="top-right" />
      {markers.map((m) => (
        <Marker
          key={m.path}
          latitude={m.lat}
          longitude={m.lng}
          anchor="bottom"
          onClick={(e) => {
            e.originalEvent.stopPropagation();
            handleMarkerClick(m);
          }}
        >
          <div className="w-6 h-6 bg-primary rounded-full border-2 border-white shadow-md cursor-pointer hover:scale-110 transition-transform" />
        </Marker>
      ))}
      {selected && (
        <Popup
          latitude={selected.lat}
          longitude={selected.lng}
          anchor="bottom"
          offset={[0, -24] as [number, number]}
          onClose={() => setSelected(null)}
          closeButton
          closeOnClick={false}
        >
          <div className="text-sm p-1">
            <button
              type="button"
              className="font-medium text-primary hover:underline"
              onClick={() => onNavigate(selected.path)}
            >
              {selected.title}
            </button>
            <div className="text-xs text-muted-foreground mt-0.5">
              {selected.lat.toFixed(4)}, {selected.lng.toFixed(4)}
            </div>
          </div>
        </Popup>
      )}
    </Map>
  );
}
