(ns meal-sched.core
  (:require [clojure.data.priority-map :refer [priority-map-by]]))

;; Invariants:
;; Each bucket maps a dimension to a bucket count
;; The sum of all bucket counts is equal to the number of resources
(def system
  (atom
   {:labels
    {:main "Main dish", :side "Side Dish", :dessert "Dessert & Drinks"},
    :dimensions #{:side :dessert :main},
    :buckets {:main 7, :side 5, :dessert 4},
    :allocations
    {"Shafer" :side,
     "Burns" :side,
     "C. Milroy" :side,
     "J. Milroy" :dessert,
     "Cassidy" :main,
     "Irwin" :main,
     "Hamer" :side,
     "A. Meredith" :main,
     "Stickel" :dessert,
     "Anderson" :main,
     "Castro" :main,
     "Reichenberg" :main,
     "S. Meredith" :dessert,
     "DeBenedittis" :dessert,
     "Friedly" :main,
     "Cowley" :side},
    :resources
    {"Shafer" {:main 1, :side 0, :dessert 3},
     "Burns" {:main 1, :side 0, :dessert 5},
     "C. Milroy" {:main 1, :side 0, :dessert 5},
     "J. Milroy" {:main 2, :side 1, :dessert 0},
     "Cassidy" {:main 0, :side 3, :dessert 1},
     "Irwin" {:main 0, :side 3, :dessert 1},
     "Hamer" {:main 1, :side 0, :dessert 3},
     "A. Meredith" {:main 0, :side 3, :dessert 1},
     "Stickel" {:main 2, :side 1, :dessert 0},
     "Anderson" {:main 0, :side 1, :dessert 3},
     "Castro" {:main 0, :side 3, :dessert 1},
     "Reichenberg" {:main 0
                    , :side 2, :dessert 3},
     "S. Meredith" {:main 2, :side 1, :dessert 0},
     "DeBenedittis" {:main 1, :side 2, :dessert 0},
     "Friedly" {:main 0, :side 1, :dessert 4},
     "Cowley" {:main 2, :side 0, :dessert 1}}}))

(defn get-distance
  "Gets the euclidian distance of a resource from the origin"
  [resource]
  (Math/sqrt (reduce + (map #(Math/pow % 2)
                            (vals resource)))))

(defn allocate
  "Allocates a resource to a bucket by resetting one dimension to the origin and incrementing the rest"
  [resource dimension]
  (into {} (->> resource
                (map (fn [[k v]]
                       [k (if (= k dimension)
                            0 (inc v))])))))

(defn get-delta
  "Gets the difference in the distance of the given resource before and after allocating it
  to a bucket in the given dimension"
  [resource dimension]
  (- (get-distance resource)
     (get-distance (allocate resource dimension))))

(defn combinations
  "Get all [resource, dimension] allocation pairs"
  [resource-keys dimensions]
  (for [k resource-keys
        d dimensions]
    [k d]))

(defn distance-pqueue
  [resources dimensions]
  (into (priority-map-by >)
        (map (fn [[k dimension]]
               (let [resource (get resources k)]
                 [[k dimension] (get-delta resource dimension)]))
             (combinations (keys resources) dimensions))))

(defn get-all-allocations [system]
  (let [{:keys [resources dimensions buckets]} system
        queue (distance-pqueue resources dimensions)]
    (loop [counts buckets
           allocs queue
           remaining resources
           allocations []]
      (if (every? #(= 0 %) (vals counts))
        allocations
        (let [allocs1 (drop-while (fn [[[key dim] _]] (or (nil? (get remaining key))
                                                          (= 0 (get counts dim)))) allocs)
              [[r-key sel-dimension] _] (first allocs1)
              resource (get remaining r-key)]
          (recur (update-in counts [sel-dimension] dec)
                 (rest allocs1)
                 (dissoc remaining r-key)
                 (conj allocations [r-key sel-dimension])))))))

(defn tx-allocate [system allocation]
  (let [[resource-key dimension] allocation]
    (-> system
        (update-in [:resources resource-key] allocate dimension)
        (update-in [:allocations] assoc resource-key dimension))))

(defn tx-allocate-all [system allocations]
  (reduce tx-allocate system allocations))

(defn tick! []
  (swap! system (fn [sys]
                  (tx-allocate-all sys (get-all-allocations sys)))))

(defn init-resource [ds]
  (reduce (fn [acc d] (assoc acc d 0)) {} ds))

(defn create-resource [sys key]
  (let [dimensions (:dimensions sys)]
    (update-in sys [:resources key]
               (constantly (init-resource dimensions)))))

(defn inc-bucket [sys bucket]
  (update-in sys [:buckets bucket] inc))

(defn add-resource! [key bucket]
  (let [dimensions (:dimensions @system)]
    (swap! system #(-> %
                       (inc-bucket bucket)
                       (create-resource key)))))

(defn print-schedule [sys]
  (let [{:keys [labels allocations]} sys]
    (doseq [[group-key items] (group-by second allocations)]
      (println (get labels group-key))
      (doseq [[resource _] (sort-by first items)]
        (println (str "\t" resource))))))

(comment 
  (print-schedule @system)
  (tick!)
  (clojure.pprint/pprint @system)
  (print-schedule @system)
  )

